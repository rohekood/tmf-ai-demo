package main

import (
"context"
"fmt"
"log"

"github.com/testcontainers/testcontainers-go"
"github.com/testcontainers/testcontainers-go/wait"
)

func main() {
req := testcontainers.ContainerRequest{
Image:        "postgres:15",
WaitingFor:   wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
}
_, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
ContainerRequest: req,
Started:          true,
})
if err != nil {
log.Fatalf("error: %v", err)
}
fmt.Println("success")
}
