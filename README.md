# goAttendance
## Step 1: Getting the required files
Make sure all rosters and attendance sheets are up to date
Open Firefox, and navigate to the gradebook
For each mod, copy the Request Payload into the matching ./postTemplates/mod*.json file, so that the request has up to date rosters
## Step 2: config.env
Various settings can be tweaked in ./config.env
A sample config can be found in ./config.env.sample

## Step 3: Run
go run main.go
