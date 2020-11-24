go 1.15

//module github.com/jdvober/goAttendance

replace github.com/jdvober/goSheets/values => ../goSheets/values/

module github.com/jdvober/goAttendance

require (
	github.com/jdvober/goGoogleAuth v0.0.0-20201015191935-8a1c594381c2
	github.com/jdvober/goSheets/values v0.0.0-00010101000000-000000000000
	github.com/joho/godotenv v1.3.0
)
