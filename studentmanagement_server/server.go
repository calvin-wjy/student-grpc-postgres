package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/crshao/go-studentmanagement-grpc/studentmanagement"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func NewStudentManagementServer() *StudentManagementServer {
	return &StudentManagementServer{}
}

type StudentManagementServer struct {
	conn *pgx.Conn
	pb.UnimplementedStudentManagementServer
}

func (server *StudentManagementServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterStudentManagementServer(s, server)
	log.Printf("Server listening at %v", lis.Addr())

	return s.Serve(lis)
}

func (server *StudentManagementServer) CreateNewStudent(ctx context.Context, in *pb.NewStudent) (*pb.Student, error) {
	log.Printf("Received: %v", in.GetName())

	createSQL := `
	CREATE TABLE IF NOT EXISTS students(
		id SERIAL PRIMARY KEY,
		name text,
		nim text
	);
	`

	_, err := server.conn.Exec(context.Background(), createSQL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Table creation failed: %v\n", err)
		os.Exit(1)
	}

	created_student := &pb.Student{Name: in.GetName(), Nim: in.GetNim()}
	tx, err := server.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("conn.Begin failed: %v", err)
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO students(name, nim) VALUES ($1, $2)", created_student.Name, created_student.Nim)
	if err != nil {
		log.Fatalf("tx.Exec failed: %v", err)
	}

	tx.Commit(context.Background())
	return created_student, nil
}

func (server *StudentManagementServer) GetStudents(ctx context.Context, in *pb.GetStudentsParams) (*pb.StudentsList, error) {
	var students_list *pb.StudentsList = &pb.StudentsList{}

	rows, err := server.conn.Query(context.Background(), "SELECT * FROM students")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		student := pb.Student{}
		err = rows.Scan(&student.Id, &student.Name, &student.Nim)
		if err != nil {
			return nil, err
		}
		students_list.Students = append(students_list.Students, &student)
	}

	return students_list, nil
}

func main() {
	database_url := "postgres://calvin.wijaya:calvin.wijaya@localhost:5432/postgres"
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Unable to establish connection: %v", err)
	}

	defer conn.Close(context.Background())

	var student_management_server *StudentManagementServer = NewStudentManagementServer()
	student_management_server.conn = conn

	if err := student_management_server.Run(); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
