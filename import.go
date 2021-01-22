package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func dbConn() (db *sql.DB) {
	er := godotenv.Load(".env")
	if er != nil {
		panic(er.Error())
	}
	dbDriver := os.Getenv("DB_Driver")
	dbUser := os.Getenv("DB_User")
	dbPass := os.Getenv("DB_Password")
	dbName := os.Getenv("DB_Name")
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

type view struct {
	Num         string
	Name        string
	Phonenumber int
	Emailid     string
	Designation string
	Country     string
	Dob         string
	Doj         string
}

type number struct {
	Totals int
}

var tmpl = template.Must(template.ParseGlob("templates/*"))

func upload(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM import ORDER BY no DESC")
	if err != nil {
		panic(err.Error())
	}
	emp := view{}
	res := []view{}
	for selDB.Next() {
		var pnum int
		var num, name, emailid, designation, country, dob, doj string
		err = selDB.Scan(&num, &name, &pnum, &emailid, &designation, &country, &dob, &doj)
		if err != nil {
			panic(err.Error())
		}
		emp.Num = num
		emp.Name = name
		emp.Phonenumber = pnum
		emp.Emailid = emailid
		emp.Designation = designation
		emp.Country = country
		emp.Dob = dob
		emp.Doj = doj
		res = append(res, emp)
	}

	defer db.Close()
	tmpl.ExecuteTemplate(w, "list.html", res)
}

func importview(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "import.html", nil)
}

func samplefile(w http.ResponseWriter, r *http.Request) {
	//Get the Excel file with in the path
	file, _ := excelize.OpenFile("uploadfiles/Sample.xlsx")
	w.Header().Set("Content-Disposition", "attachment; filename="+string("Sample.xlsx")+"")
	file.Write(w)

}

func importdata(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	r.ParseMultipartForm(200000)
	formdata := r.MultipartForm
	fil := formdata.File["files"] // grab the filenames
	for i := range fil {          // loop through the files one by one

		//file save to open
		file, err := fil[i].Open()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		defer file.Close()

		// fname := fil[i].Filename

		tempFile, err := ioutil.TempFile("uploadfiles", "upload-*.xlsx")
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()
		filepath := tempFile.Name()
		log.Println("tempfile is :", tempFile.Name())

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)

		//Select and open Excel File

		f, err := excelize.OpenFile(filepath)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get all the rows in the vegan section.
		rows, err := f.GetRows("Sheet1")
		i := 1
		for _, row := range rows {

			if i > 1 {

				var temp1 view
				var str []string
				for _, colCell := range row {
					str = append(str, colCell)
					fmt.Println("str append is :", str)
				}

				// no := str[0]
				temp1.Num = str[0]
				temp1.Name = str[1]
				pno := str[2]
				temp1.Phonenumber, _ = strconv.Atoi(pno)
				temp1.Emailid = str[3]
				temp1.Designation = str[4]
				temp1.Country = str[5]
				temp1.Dob = str[6]
				temp1.Doj = str[7]
				db := dbConn()
				stmt, _ := db.Prepare("insert into import(Name,PhoneNumber,Emailid,Designation,Country,Dob,Doj) values(?,?,?,?,?,?,?)") //Get prepared statement object

				fmt.Println(temp1.Num, temp1.Name, temp1.Phonenumber, temp1.Emailid, temp1.Designation, temp1.Country, temp1.Dob, temp1.Doj)
				stmt.Exec(temp1.Name, temp1.Phonenumber, temp1.Emailid, temp1.Designation, temp1.Country, temp1.Dob, temp1.Doj) //Call prepared statement
				defer db.Close()

			}
			i++
		}
	}

	http.Redirect(w, r, "/", 301)

}

func exportdata(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	f := excelize.NewFile()
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM import")
	if err != nil {
		panic(err.Error())
	}

	inc := 1
	for selDB.Next() {
		inc = inc + 1
		incstr := strconv.Itoa(inc)
		var pnum int
		var num, name, emailid, designation, country, dob, doj string
		err = selDB.Scan(&num, &name, &pnum, &emailid, &designation, &country, &dob, &doj)
		if err != nil {
			panic(err.Error())
		}
		f.SetCellValue("Sheet1", "a1", "ID")
		f.SetCellValue("Sheet1", "b1", "Name")
		f.SetCellValue("Sheet1", "c1", "PhoneNumber")
		f.SetCellValue("Sheet1", "d1", "EmailId")
		f.SetCellValue("Sheet1", "e1", "Designtion")
		f.SetCellValue("Sheet1", "f1", "Country")
		f.SetCellValue("Sheet1", "g1", "Date of birt")
		f.SetCellValue("Sheet1", "h1", "Date of joining")
		f.SetCellValue("Sheet1", "a"+incstr, num)
		f.SetCellValue("Sheet1", "B"+incstr, name)
		f.SetCellValue("Sheet1", "c"+incstr, pnum)
		f.SetCellValue("Sheet1", "d"+incstr, emailid)
		f.SetCellValue("Sheet1", "e"+incstr, designation)
		f.SetCellValue("Sheet1", "f"+incstr, country)
		f.SetCellValue("Sheet1", "g"+incstr, dob)
		f.SetCellValue("Sheet1", "h"+incstr, doj)
	}

	if err = f.SaveAs("uploadfiles/Export.xlsx"); err != nil {
		println(err.Error())
	}

	//Get the Excel file with in the path
	file, err := excelize.OpenFile("uploadfiles/Export.xlsx")
	w.Header().Set("Content-Disposition", "attachment; filename="+string("Export.xlsx")+"")
	file.Write(w)
	// io.Copy(w, file)
	// time.Sleep(1 * time.Second)
	// http.Redirect(w, r, "/", 301)
}

func count(w http.ResponseWriter, r *http.Request) {
	totl := number{}
	res1 := []number{}
	r.ParseMultipartForm(200000)
	formdata := r.MultipartForm
	fil := formdata.File["files"] // grab the filenames
	for i := range fil {          // loop through the files one by one

		//file save to open
		file, err := fil[i].Open()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		defer file.Close()

		// fname := fil[i].Filename

		tempFile, err := ioutil.TempFile("uploadfiles", "upload-*.xlsx")
		if err != nil {
			fmt.Println(err)
		}
		defer tempFile.Close()
		filepath := tempFile.Name()
		log.Println("tempfile is :", tempFile.Name())

		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)

		//Select and open Excel File

		f, err := excelize.OpenFile(filepath)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get all the rows in the vegan section.
		rows, err := f.GetRows("Sheet1")

		fmt.Println("Data in row", rows)

		num := len(rows)
		totl.Totals = num - 1

		res1 = append(res1, totl)

		fmt.Println("number of rows", totl.Totals)
	}
	// m := map[string]interface{}{

	// 	"Total": res1,
	// }
	count := strconv.Itoa(totl.Totals)
	w.Write([]byte(count))
}

func main() {
	log.Println("Server started on: http://localhost:7000")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	// http.Handle("/img/", http.StripPrefix("/img", http.FileServer(http.Dir("img"))))
	// http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("css"))))
	http.HandleFunc("/", upload)
	http.HandleFunc("/importview", importview)
	http.HandleFunc("/importdata", importdata)
	http.HandleFunc("/samplefile", samplefile)
	http.HandleFunc("/exportdata", exportdata)
	http.HandleFunc("/count", count)
	http.ListenAndServe(":7000", nil)
}
