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



func indexerEngine(rootPath string) {
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

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error accessing path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jpeg") {
			fileData, err := extractText(path)
			if err != nil {
				fmt.Printf("error extracting text from %q: %v\n", path, err)
				return err
			}

			file := File{
				FileName: info.Name(),
				FilePath: path,
				FileData: fileData,
			}
			// Log the entire file struct to the terminal
            fmt.Printf("Processed File: %+v\n", file)
			// Insert the File data into MongoDB
			_, err = collection.InsertOne(context.TODO(), file)
			if err != nil {
				fmt.Printf("error inserting data into MongoDB: %v\n", err)
				return err
			}
			 // Log successful saving to the terminal
			 fmt.Println("Successfully saved to MongoDB:", file.FileName)
		}
		return nil // Important to return nil if no error occurred
	})

	if err != nil {
		fmt.Printf("error walking the path %v: %v\n", rootPath, err)
	}
}


