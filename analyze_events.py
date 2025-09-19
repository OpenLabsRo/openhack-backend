import json

with open('/home/sunshine/coding/openlabs/openhack-backend/events_dump.json', 'r') as f:
    try:
        data = json.load(f)
    except json.JSONDecodeError as e:
        print(f"Error decoding JSON: {e}")
        f.seek(0)
        data = []
        for line in f:
            try:
                data.append(json.loads(line))
            except json.JSONDecodeError:
                pass

event_counts = {}
for event in data:
    action = event.get('action')
    if action:
        event_counts[action] = event_counts.get(action, 0) + 1

for action, count in sorted(event_counts.items()):
    print(f"{action}: {count}")
