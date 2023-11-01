package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"log"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"context"
)

type File struct {
	FileName string
	FilePath string
	FileData string
}

func saveFilesToDB(files []File) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in environment")
	}
	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	// Get a handle for your collection
	collection := client.Database("tesseract").Collection("images")

	// Insert multiple documents
	var insertData []interface{}
	for _, file := range files {
		insertData = append(insertData, file)
	}
	
	_, err = collection.InsertMany(context.TODO(), insertData)
	if err != nil {
		log.Fatal(err)
	}

	// Close the connection
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

func folderCrawler(rootPath string) []File {
	var jpegFiles []File
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error accessing path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jpeg") {
			fileData,_ := extractText(path)
			jpegFiles = append(jpegFiles, File{
				FileName: info.Name(),
				FilePath: path,
				FileData: fileData,
			})
		}
		

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %v: %v\n", rootPath, err)
	}
	return jpegFiles
}

func indexerEngine(rootPath string){
	jpegFiles:=folderCrawler(rootPath)
	saveFilesToDB(jpegFiles)
}
