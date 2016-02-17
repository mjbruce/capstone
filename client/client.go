package main

import (
	"fmt"
	"io"
	"os"
	"net"
	"bufio"
	"strings"
	"strconv"
	"path/filepath"
	"errors"
)

type Peer struct {
	IP   string
	Port string
}

type File struct {
	length       int
	path         string
	name         string
	pieces       string
	piece_length int
}

var files []File
var trackerIP string // Will be set after parsing Metainfo
var peers []Peer

const END_OF_ENTRY = ":#!"
const META_VALUE_INDEX = 1

func deleteEntry(nameToDelete string) {

	i := 0
	for i < len(files){
		if(nameToDelete == files[i].name){
			if len(files) > 2{
				files = append(files[:i], files[i+1:]...)
			}else if(i == 0){
				files = append(files[i:])
			}else if(i == 1){
				files = append(files[:i])
			}
		}
		i++
	}
}

func updateMetainfo() error {

	err := os.Remove("meta.info")
	if err != nil{
		fmt.Print(err)
		return err
	}

	newMetainfo, err := os.Create("meta.info")
	if err != nil {
		//fmt.Println("hi")
		return err
	}

	i := 0
	newMetainfo.WriteString("annouce:::" + trackerIP + "\n")
	for i < len(files) {

		newMetainfo.WriteString("length:::" + strconv.Itoa(files[i].length) +"\n")
		newMetainfo.WriteString("path:::" + files[i].path + "\n")
		newMetainfo.WriteString("name:::" + files[i].name + "\n")
		newMetainfo.WriteString("pieces_length:::" + strconv.Itoa(files[i].piece_length) + "\n")
		newMetainfo.WriteString("pieces:::" + files[i].pieces + "\n")
		newMetainfo.WriteString(END_OF_ENTRY + "\n")
		i++

	}

	return newMetainfo.Close()

}

func parseMetainfo(metainfo_path string) error {
	metainfo_file, err := os.Open(metainfo_path)
	if err != nil {
		return err
	} else if metainfo_path != "meta.info" {
		return errors.New("Invalid File Type")
	}

	scanner := bufio.NewScanner(metainfo_file)
	temp_file := File{}

	for scanner.Scan() {

		line  := strings.TrimSpace(scanner.Text())
		split := strings.Split(line, ":::")

		if split[0] == "announce" {
			trackerIP = split[META_VALUE_INDEX]
		} else if split[0] == "pieces_length" {
			temp_int, err := strconv.Atoi(split[META_VALUE_INDEX])
			temp_file.piece_length = temp_int
			if err != nil {
				return err
			}
		} else if split[0] == "length" {
			temp_int, err := strconv.Atoi(split[META_VALUE_INDEX])
			temp_file.length = temp_int
			if err != nil {
				return err
			}
		} else if (strings.Contains(line, "path")) {
			temp_file.path = split[META_VALUE_INDEX]
		} else if (strings.Contains(line, "name")) {
			temp_file.name = split[META_VALUE_INDEX]
		} else if (strings.Contains(line, "pieces")) {
			temp_file.pieces = split[META_VALUE_INDEX]
		} else if (strings.Contains(line, END_OF_ENTRY)) {
			files = append(files, temp_file)  //if there is a duplicate whole thing goes empty
			temp_file = File{}
		}

	}

	//fmt.Printf("%v", files)

	return metainfo_file.Close()
}

func addToMetainfo(path_add, path_metainfo string) error {
	//adding_file, err := os.Open(path_add)
	metainfo_file, err := os.OpenFile(path_metainfo, os.O_APPEND | os.O_WRONLY, 0644)
																		//appends to metainfo
																		// needs permissions?
	if err != nil{
		return err
	}

	add_info,err := os.Stat(path_add)
	if err != nil{
		return err
	}

	parseMetainfo(path_metainfo)
	i := 0
	for i < len(files) {
		if files[i].name == add_info.Name() {
			//fmt.Println(files[i].name)
			return errors.New("Can't Add Duplicates To Metainfo")
		}
		i++
	}

	temp_size := add_info.Size() 			//write length
	temp_str := strconv.FormatInt(temp_size,10)
	metainfo_file.WriteString("length:::" + temp_str + "\n")

	temp_file_path, err := filepath.Abs(path_add)
	if err != nil{
		return err
	}

	metainfo_file.WriteString("path:::" + temp_file_path + "\n")
	metainfo_file.WriteString("name:::" + add_info.Name() + "\n")
	metainfo_file.WriteString("pieces_length:::-1\n")
	metainfo_file.WriteString("pieces:::chuncking not currently implemented\n")
	metainfo_file.WriteString(END_OF_ENTRY + "\n")

	return metainfo_file.Close()
}

func fileCopy(src, dst string) error {
	in, err := os.Open(src) // Opens input
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst) // Opens output
	if err != nil {
		return err
	}
	//defer out.Close()

	_, err = io.Copy(out, in) // Copies the file contents
	if err != nil {
		return err
	}

	cerr := out.Close() // Checks for close error
	return cerr
}



func main() {
	/*fmt.Println("Hello World!")
	fmt.Println("Cool Beans!")
	err := fileCopy(os.Args[1], os.Args[2])

	if err != nil {
		fmt.Println("You suck")
	} else {
		fmt.Println(os.Args[1] + " copied to " + os.Args[2])
	}*/

	//parseMetainfo(os.Args[1])
	parseMetainfo("meta.info")
	i := 0
	for i < len(files) {
		fmt.Println(files[i])
		if files[i].name == "test.txt" {
		}
		i++
	}
}

// ------------------------- CODE BELOW THIS LINE IS UNTESTED AND DANGEROUS ------------------------- \\

func askTrackerForPeers() {
	// Connets to tracker
	conn, err := net.Dial("tcp", trackerIP);
	if err != nil {
		return
	}

	fmt.Fprintf(conn, "Announce_Request: <Stuff>")

	reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline

	for err != nil {
		peerArray := strings.Split(reply, ":::")
		peers = append(peers, Peer{IP:peerArray[0], Port:peerArray[1]})
		reply, err = bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline
	}

}

func getFile(fileName string) {

	i := 0
	gotFile := false
	for i < len(peers) && !gotFile {
		conn, err := net.Dial("tcp", peers[i].IP);
		if err != nil {
			return
		}

		fmt.Fprintf(conn, "Do_You_Have_FileName:" + fileName)

		reply, err := bufio.NewReader(conn).ReadString('\n') // Waits for a String ending in newline

		// Has file and no errors
		if reply != "NO" && err == nil {
			file, err := os.Create(fileName)
			if err != nil {
				break // could set boolean instead
			}
			defer file.Close();

			n, err := io.Copy(conn, file)
			if err != nil {
				break // could set boolean instead
			}
			fmt.Println(n, "this was sent")
			gotFile = true
		}
		i++
	}

}

