package testslop

// ProcessCommand demonstrates a giant switch statement anti-pattern
func ProcessCommand(cmd string) string {
switch cmd {
case "start":
return "Starting service"
case "stop":
return "Stopping service"
case "restart":
return "Restarting service"
case "status":
return "Checking status"
case "deploy":
return "Deploying application"
case "rollback":
return "Rolling back deployment"
case "scale":
return "Scaling service"
case "migrate":
return "Running migrations"
case "backup":
return "Creating backup"
case "restore":
return "Restoring from backup"
case "health":
return "Health check"
case "metrics":
return "Gathering metrics"
default:
return "Unknown command"
}
}

// CategorizeValue demonstrates a giant if-else chain anti-pattern
func CategorizeValue(x int) string {
if x == 1 {
return "one"
} else if x == 2 {
return "two"
} else if x == 3 {
return "three"
} else if x == 4 {
return "four"
} else if x == 5 {
return "five"
} else if x == 6 {
return "six"
} else if x == 7 {
return "seven"
} else if x == 8 {
return "eight"
} else if x == 9 {
return "nine"
} else if x == 10 {
return "ten"
} else if x == 11 {
return "eleven"
} else {
return "other"
}
}

// HandleType demonstrates a giant type switch anti-pattern
func HandleType(v interface{}) string {
switch v.(type) {
case int:
return "integer"
case string:
return "string"
case bool:
return "boolean"
case float64:
return "float"
case []int:
return "int slice"
case []string:
return "string slice"
case map[string]int:
return "string-int map"
case map[int]string:
return "int-string map"
case struct{}:
return "struct"
case *int:
return "int pointer"
case *string:
return "string pointer"
case chan int:
return "int channel"
default:
return "unknown type"
}
}
