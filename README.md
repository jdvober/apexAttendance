# apexAttendance
## Step 1: Getting the required files
1. Make sure all rosters and attendance sheets are up to date.  Everything on Google Sheets should be set to "Plain Text".
2. Delete any students who are not on Sunguard Roster.
3. Open Firefox, and navigate to the gradebook.
4. For each mod, copy the Request Payload into the matching ./postTemplates/mod*.json file, so that the request has up to date rosters
## Step 2: config.env
Various settings can be tweaked in ./config.env
A sample config can be found in ./config.env.sample

## Step 3: Run
go run main.go
