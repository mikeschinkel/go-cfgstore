1. In examples/basic_usage/main.go, replace the heading underline's fmt.Println("===...\n") with fmt.Println("===...") followed by fmt.Println(), keeping the printed output identical.
2. Run go vet ./... and go build ./... in examples/basic_usage; both pass.
3. Run the example and confirm its output is unchanged.