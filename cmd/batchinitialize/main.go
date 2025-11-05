package main

import (
	"backend/internal/db"
	"backend/internal/env"
	"backend/internal/errmsg"
	"backend/internal/models"
	"backend/internal/utils"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ParticipantRecord struct {
	FirstName           string
	LastName            string
	University          string
	DateOfBirth         string
	TeamName            string
	PhoneNumber         string
	MedicalConditions   string
	DietaryRestrictions string
	Email               string
}

var (
	deploymentFlag = flag.String("deployment", "dev", "deployment profile (dev|test|prod)")
	envRoot        = flag.String("env-root", "", "directory containing environment files")
	csvFile        = flag.String("csv", "", "path to CSV file with participant data")
)

func main() {
	flag.Parse()

	if *csvFile == "" {
		log.Fatal("CSV file path is required. Use -csv <path>")
	}

	// Initialize environment
	env.Init(*envRoot, "")

	// Initialize database
	if err := db.InitDB(*deploymentFlag); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Client.Disconnect(db.Ctx)

	// Initialize cache
	if err := db.InitCache(*deploymentFlag); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Read CSV file
	records, err := readCSV(*csvFile)
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		log.Fatal("CSV file is empty")
	}

	fmt.Printf("Processing %d participants...\n", len(records))

	// Group participants by team
	teamMap := make(map[string][]ParticipantRecord)
	for _, record := range records {
		teamName := strings.TrimSpace(record.TeamName)
		if teamName == "" {
			teamName = fmt.Sprintf("Team-%s", utils.GenID(4))
		}
		teamMap[teamName] = append(teamMap[teamName], record)
	}

	// Create teams and add participants
	successCount := 0
	failureCount := 0

	for teamName, participants := range teamMap {
		teamID, err := createTeamWithParticipants(teamName, participants)
		if err != nil {
			fmt.Printf("❌ Failed to create team '%s': %v\n", teamName, err)
			failureCount += len(participants)
			continue
		}

		fmt.Printf("✓ Created team '%s' (ID: %s) with %d participants\n", teamName, teamID, len(participants))
		successCount += len(participants)
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total teams created: %d\n", len(teamMap))
	fmt.Printf("Total participants created: %d\n", successCount)
	if failureCount > 0 {
		fmt.Printf("Failed participants: %d\n", failureCount)
	}
}

func readCSV(filePath string) ([]ParticipantRecord, error) {
	// Check if file exists
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column indices
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[strings.TrimSpace(col)] = i
	}

	// Validate required columns
	requiredColumns := []string{"FirstName", "LastName", "Email", "Team Name"}
	for _, col := range requiredColumns {
		if _, exists := columnMap[col]; !exists {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var records []ParticipantRecord

	// Read data rows
	lineNumber := 2
	for {
		line, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error reading CSV at line %d: %w", lineNumber, err)
		}

		// Helper to safely get column value
		getColumn := func(name string) string {
			if idx, ok := columnMap[name]; ok && idx < len(line) {
				return strings.TrimSpace(line[idx])
			}
			return ""
		}

		// Skip empty rows
		if getColumn("FirstName") == "" && getColumn("LastName") == "" && getColumn("Email") == "" {
			continue
		}

		record := ParticipantRecord{
			FirstName:           getColumn("FirstName"),
			LastName:            getColumn("LastName"),
			University:          getColumn("University"),
			DateOfBirth:         getColumn("Date of Birth"),
			TeamName:            getColumn("Team Name"),
			PhoneNumber:         getColumn("Phone Number"),
			MedicalConditions:   getColumn("Medical conditions"),
			DietaryRestrictions: getColumn("Dietary Restrictions"),
			Email:               getColumn("Email"),
		}

		// Validate required fields
		if (record.FirstName == "" && record.LastName == "") || record.Email == "" {
			fmt.Printf("⚠ Skipping row %d: missing FirstName/LastName or Email\n", lineNumber)
			lineNumber++
			continue
		}

		records = append(records, record)
		lineNumber++
	}

	return records, nil
}

func createTeamWithParticipants(teamName string, participants []ParticipantRecord) (string, error) {
	// Create team with custom ID and name
	team := &models.Team{
		ID:      utils.GenTeamID(),
		Name:    teamName,
		Members: []string{},
		Deleted: false,
	}

	// Insert team directly into database
	_, err := db.Teams.InsertOne(db.Ctx, team)
	if err != nil {
		return "", fmt.Errorf("failed to create team: %w", err)
	}

	// Create participants and add to team
	for i, record := range participants {
		// Limit team to 4 members
		if i >= 4 {
			fmt.Printf("  ⚠ Team '%s' reached max members (4), skipping additional participants\n", teamName)
			break
		}

		account, err := createAccountFromRecord(record, team.ID)
		if err != nil {
			fmt.Printf("  ❌ Failed to create account for %s %s: %v\n", record.FirstName, record.LastName, err)
			continue
		}

		team.Members = append(team.Members, account.ID)
		fmt.Printf("  ✓ Added %s %s (%s)\n", record.FirstName, record.LastName, account.Email)
	}

	// Update team with members using the model method
	if len(team.Members) > 0 {
		err = team.ChangeMembers(team.Members)
		if err != nil {
			return team.ID, fmt.Errorf("failed to update team members: %w", err)
		}
	}

	return team.ID, nil
}

func convertDateFormat(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse YYYY-MM-DD format
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// If parsing fails, return original string
		return dateStr
	}

	// Format to DD.MM.YYYY
	return parsedTime.Format("02.01.2006")
}

func createAccountFromRecord(record ParticipantRecord, teamID string) (*models.Account, error) {
	// Create account with all required fields
	account := &models.Account{
		Email:             record.Email,
		FirstName:         record.FirstName,
		LastName:          record.LastName,
		University:        record.University,
		DOB:               convertDateFormat(record.DateOfBirth),
		PhoneNumber:       record.PhoneNumber,
		MedicalConditions: record.MedicalConditions,
		FoodRestrictions:  record.DietaryRestrictions,
		TeamID:            teamID,
		CheckedIn:         false,
		Present:           false,
		HasVoted:          false,
		Consumables: models.Consumables{
			Water:      0,
			Pizza:      false,
			Coffee:     false,
			Jerky:      false,
			Sandwiches: 0,
		},
	}

	// Use Initialize method which generates unique ID and inserts into database
	serr := account.Initialize()
	if serr != errmsg.EmptyStatusError {
		return nil, fmt.Errorf("failed to initialize account: %s", serr.Message)
	}

	return account, nil
}
