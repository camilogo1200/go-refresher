package greetings

import "fmt"
	
func Hello(name string) string {
	
	nameText := ""
	template := ""
	if len(name) <= 0 {
		nameText = ""
		template = "Hi & Welcome! %v"
	} else {
		nameText = name
		template = "Hi, %v. Welcome!"
	}
	
	//Return a greeting that embeds the name in a message
	message := fmt.Sprintf(template, nameText);
	return message;
}
