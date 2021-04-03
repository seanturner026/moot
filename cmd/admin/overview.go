package main

func (app application) overviewHandler() (string, int) {
	message := "User authorized successfully"
	statusCode := 200
	return message, statusCode
}
