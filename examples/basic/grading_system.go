package main

import (
	"fmt"

	"github.com/jongyunha/shrinkmap"
)

// Combined example showing both batch operations and iterator usage
func gradingSystemExample() {
	sm := shrinkmap.New[string, float64](shrinkmap.DefaultConfig())
	defer sm.Stop()

	// Add initial grades using batch operation
	initialGrades := shrinkmap.BatchOperations[string, float64]{
		Operations: []shrinkmap.BatchOperation[string, float64]{
			{Type: shrinkmap.BatchSet, Key: "student1", Value: 85.5},
			{Type: shrinkmap.BatchSet, Key: "student2", Value: 92.0},
			{Type: shrinkmap.BatchSet, Key: "student3", Value: 78.5},
			{Type: shrinkmap.BatchSet, Key: "student4", Value: 95.0},
		},
	}

	if err := sm.ApplyBatch(initialGrades); err != nil {
		fmt.Printf("Failed to add initial grades: %v\n", err)
		return
	}

	// Update grades after final exam using batch operation
	finalExamUpdates := shrinkmap.BatchOperations[string, float64]{
		Operations: []shrinkmap.BatchOperation[string, float64]{
			{Type: shrinkmap.BatchSet, Key: "student1", Value: 88.0},
			{Type: shrinkmap.BatchSet, Key: "student2", Value: 94.5},
			{Type: shrinkmap.BatchSet, Key: "student3", Value: 82.0},
			{Type: shrinkmap.BatchSet, Key: "student4", Value: 97.5},
		},
	}

	if err := sm.ApplyBatch(finalExamUpdates); err != nil {
		fmt.Printf("Failed to update final grades: %v\n", err)
		return
	}

	// Use iterator to generate final report
	iter := sm.NewIterator()
	fmt.Println("\nFinal Grade Report:")
	fmt.Println("==================")

	var totalGrade float64
	var highestGrade float64
	var highestStudent string
	count := 0

	for iter.Next() {
		student, grade := iter.Get()
		fmt.Printf("%s: %.1f\n", student, grade)

		totalGrade += grade
		count++

		if grade > highestGrade {
			highestGrade = grade
			highestStudent = student
		}
	}

	if count > 0 {
		fmt.Println("\nClass Statistics:")
		fmt.Printf("Average Grade: %.1f\n", totalGrade/float64(count))
		fmt.Printf("Highest Grade: %.1f (Student: %s)\n", highestGrade, highestStudent)
	}
}
