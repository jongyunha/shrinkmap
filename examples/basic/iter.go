package main

import (
	"fmt"

	"github.com/jongyunha/shrinkmap"
)

// IteratorExample demonstrates various iterator use cases
func iteratorExample() {
	sm := shrinkmap.New[string, float64](shrinkmap.DefaultConfig())
	defer sm.Stop()

	// Add some student grades
	grades := map[string]float64{
		"Alice":   95.5,
		"Bob":     87.0,
		"Charlie": 92.5,
		"David":   88.5,
		"Eve":     94.0,
	}

	for student, grade := range grades {
		_ = sm.Set(student, grade)
	}

	// Use iterator to calculate average grade
	iter := sm.NewIterator()
	var sum float64
	count := 0

	fmt.Println("Student Grades:")
	for iter.Next() {
		student, grade := iter.Get()
		fmt.Printf("%s: %.1f\n", student, grade)
		sum += grade
		count++
	}

	if count > 0 {
		average := sum / float64(count)
		fmt.Printf("\nClass Average: %.1f\n", average)
	}

	// Using iterator to find top performers (grade >= 90)
	iter = sm.NewIterator() // Create new iterator for second pass
	fmt.Println("\nTop Performers:")
	for iter.Next() {
		student, grade := iter.Get()
		if grade >= 90 {
			fmt.Printf("%s: %.1f\n", student, grade)
		}
	}
}
