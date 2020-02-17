package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

// Student model
type Student struct {
	ID         int
	Name       string `gorm:"type:varchar(100);unique;not null"`
	Class      string `gorm:"type:varchar(100);not null"`
	Enrollment string `gorm:"type:bigint(8);unique;not null;"`
}

func main() {
	reset()
	setup()
	seed()
	options()
}

func clear() {
	fmt.Print("\033[H\033[2J")
}

func reset() {
	db, _ = gorm.Open("mysql", "root:@(127.0.0.1)/studentsDB?charset=utf8&parseTime=True&loc=Local")
	db.DropTableIfExists(&Student{})
}
func setup() {
	var err error
	db, err = gorm.Open("mysql", "root:@(127.0.0.1)/studentsDB?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	if !db.HasTable(&Student{}) {
		db.CreateTable(&Student{})
	}
	// defer db.Close()
	db.AutoMigrate(&Student{})
}

func options() {
	var input string
	for input != "5" {
		input = "0"
		clear()
		fmt.Println("1) Create student")
		fmt.Println("2) Update student")
		fmt.Println("3) Delete student")
		fmt.Println("4) Search student")
		fmt.Println("5) Exit")
		fmt.Print("Select an option: ")
		fmt.Scanln(&input)
		switch input {
		case "1":
			clear()
			create()
		case "2":
			clear()
			update()
		case "3":
			clear()
			delete()
		case "4":
			clear()
			fmt.Print("Type a name to search(empty to show all): ")
			var param string
			fmt.Scanln(&param)
			find(param)
			pressForContinue()
		case "5":
			clear()
		}
	}
}

func createStudent(student Student) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&student).Error; err != nil {
			return err
		}
		return nil
	})
}

func updateStudent(student Student) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&student).Error; err != nil {
			return err
		}
		return nil
	})
}

func deleteStudent(student Student) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&student).Error; err != nil {
			return err
		}
		return nil
	})
}

func create() {
	var name, class, enrollment string
	fmt.Print("Type the student's name: ")
	fmt.Scanln(&name)
	fmt.Print("Type the student's class: ")
	fmt.Scanln(&class)
	fmt.Print("Type the student's enrollment: ")
	fmt.Scanln(&enrollment)
	student := Student{Name: name, Class: strings.ToUpper(class), Enrollment: enrollment}
	err := createStudent(student)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Added successfully")
	pressForContinue()
}

func delete() {
	var id string
	fmt.Print("Enter the id of the student you want to delete: ")
	fmt.Scanln(&id)
	intID, _ := strconv.Atoi(id)
	founded, student := findByID(intID)
	if founded {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Are you sure you want to delete this record? (y): ")
		if scanner.Scan() {
			if str := scanner.Text(); str == "y" || str == "" {
				err := deleteStudent(student)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Deleted successfully")
				}
				// fmt.Println("Deleted successfully")
				pressForContinue()
			} else {
				fmt.Println("Delete aborted")
				pressForContinue()
			}
		}
	} else {
		fmt.Println("No student found")
		pressForContinue()
	}
}

func update() {
	var id string
	fmt.Print("Enter the id of the student you want to update: ")
	fmt.Scanln(&id)
	intID, _ := strconv.Atoi(id)
	founded, student := findByID(intID)
	if founded {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Type the student's name(empty to keep '", student.Name, "'): ")
		if scanner.Scan() {
			if str := scanner.Text(); len(str) != 0 {
				student.Name = str
			}
		}
		fmt.Print("Type the student's class(empty to keep '", student.Class, "'): ")
		if scanner.Scan() {
			if str := scanner.Text(); len(str) != 0 {
				student.Class = strings.ToUpper(str)
			}
		}
		fmt.Print("Type the student's enrollment(empty to keep '", student.Enrollment, "'): ")
		if scanner.Scan() {
			if str := scanner.Text(); len(str) != 0 {
				student.Enrollment = str
			}
		}
		err := updateStudent(student)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Updated succesfully")
		pressForContinue()
	} else {
		fmt.Println("No student found")
		pressForContinue()
	}
}

func findStudentsByID(id int, student Student) (Student, error) {
	return student, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&student, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func findByID(id int) (bool, Student) {
	var student Student
	student, err := findStudentsByID(id, student)
	if err != nil && err.Error() != "record not found" {
		fmt.Println(err)
	}
	if student.ID != 0 {
		printStudents([]Student{student})
		return true, student
	}
	return false, Student{}
}

func findStudents(param string, students []Student) ([]Student, error) {
	return students, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("name LIKE ?", "%"+param+"%").Find(&students).Error; err != nil {
			return err
		}
		return nil
	})
}

func getAllStudents(students []Student) ([]Student, error) {
	return students, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Find(&students).Error; err != nil {
			return err
		}
		return nil
	})
}

func find(param string) {
	var students []Student
	var err error
	if len(param) == 0 {
		students, err = getAllStudents(students)
	} else {
		students, err = findStudents(param, students)
	}
	if err != nil {
		fmt.Println(err)
	}
	printStudents(students)
}

func printStudents(students []Student) {
	if len(students) != 0 {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
		fmt.Println("STUDENTS")
		fmt.Fprintln(w, "\t  ID\t  NAME\t  ENROLLMENT\t  CLASS\t")
		for _, student := range students {
			fmt.Fprintln(w, ("\t  " + strconv.Itoa(student.ID) + "\t  " + student.Name + "\t  " + student.Enrollment + "\t  " + student.Class + "\t"))
		}
		w.Flush()
	} else {
		fmt.Println("No records found")
	}
}

func pressForContinue() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Press enter to continue")
	scanner.Scan()
}

func seed() {
	db.Save(&Student{Name: "Alfredo Aragon", Class: "ITIC 10-1", Enrollment: "19104001"})
	db.Save(&Student{Name: "Jose Noriega", Class: "ITIC 10-1", Enrollment: "19104002"})
	db.Save(&Student{Name: "Abiel Robledo", Class: "ITIC 10-1", Enrollment: "19104003"})
}
