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
	"sync"
)


const MaxWorkers = 16

func worker(filesChan <-chan string, wg *sync.WaitGroup, collection *mongo.Collection) {
	defer wg.Done()

	for path := range filesChan {
		fileData, err := extractText(path)
		if err != nil {
			fmt.Printf("error extracting text from %q: %v\n", path, err)
			continue
		}

		fileName := filepath.Base(path)
		file := File{
			FileName: fileName,
			FilePath: path,
			FileData: fileData,
		}
		// Log the entire file struct to the terminal
		// fmt.Printf("Processed File: %+v\n", file)

		// Insert the File data into MongoDB
		_, err = collection.InsertOne(context.TODO(), file)
		if err != nil {
			fmt.Printf("error inserting data into MongoDB: %v\n", err)
			continue
		}
		// Log successful saving to the terminal
		// fmt.Println("Successfully saved to MongoDB:", file.FileName)
	}
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

	filesChan := make(chan string, MaxWorkers)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go worker(filesChan, &wg, collection)
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error accessing path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jpeg") {
			filesChan <- path
		}
		return nil
	})

	close(filesChan)  // Close the channel when file listing is done
	wg.Wait()         // Wait for all workers to finish

	if err != nil {
		fmt.Printf("error walking the path %v: %v\n", rootPath, err)
	}
}
