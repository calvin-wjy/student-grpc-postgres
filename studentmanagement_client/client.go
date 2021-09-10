package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/crshao/go-studentmanagement-grpc/studentmanagement"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	c := pb.NewStudentManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var new_students = make(map[string]string)
	new_students["Calvin"] = "123456"
	new_students["Ali"] = "123123"

	for name, nim := range new_students {
		r, err := c.CreateNewStudent(ctx, &pb.NewStudent{Name: name, Nim: nim})

		if err != nil {
			log.Fatalf("Could not create student: %v", err)
		}

		log.Printf(`Student Details:
NAME: %s
NIM: %s
ID: %d`, r.GetName(), r.GetNim(), r.GetId())
	}

	params := &pb.GetStudentsParams{}
	r, err := c.GetStudents(ctx, params)

	if err != nil {
		log.Fatalf("could not retrieve students: %v", err)
	}

	log.Print("\nSTUDENT LIST:\n")
	fmt.Printf("r.GetStudents(): %v\n", r.GetStudents())
}
