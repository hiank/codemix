package main

import (
	"bytes"
	"strings"
	"io/ioutil"
	"github.com/hiank/core"
	"github.com/hiank/codemix"
	"os"
	"fmt"
	"flag"
)


func main() {

	srcDir := flag.String("s", "", "src dir")
	dstDir := flag.String("d", "", "dst dir")
	// mixLen := flag.Int("l", 300, "the num of the mix byte")

	flag.Parse()

	// *srcDir = "F:\\Classes_"
	// *dstDir = "F:\\tmp"
	if *srcDir == "" {
		fmt.Println("必须通过 -s 指定源文件目录")
		return
	}

	if *dstDir == "" {
		*dstDir = *srcDir
	}

// 	fmt.Println("s:" + *srcDir + "___d:" + *dstDir)
// 	if *srcDir == *dstDir {

// 		fmt.Print("检测到目标目录与源目录一样，此操作将覆盖源文件，是(Y)否(N)继续？：")
// 		reader, goon := bufio.NewReader(os.Stdin), false
// L:		for {
// 			data, _ := reader.ReadByte()
// 			switch data {
// 				case 'Y': goon = true; break L
// 				case 'N': goon = false; break L
// 				default:
// 			}
// 		}
// 		if !goon {
// 			fmt.Println("尝试使用 -d 指定目标目录")		
// 			return
// 		}
// 	}

	// t := []byte(*dstDir)
	// apart := []byte{core.DirApartByte()}
	apartStr := string(os.PathSeparator)
	fmt.Println("dstDir is : " + *dstDir)
	if !strings.HasSuffix(*dstDir, apartStr) {
		// t := t[1:]
		*dstDir += apartStr
		fmt.Println("dstDir is : " + *dstDir)
	}
	
	loadFile(srcDir, func (h []byte, cpp []byte, fileName string) int {

		hasH := false
		switch {
			case strings.HasSuffix(fileName, ".cpp"): 
				h, cpp = codemix.Mix(h, cpp); fmt.Println(fileName + "__OK")
				hasH = true
			// case strings.HasSuffix(fileName, ".h"):
			case strings.HasSuffix(fileName, ".mm"): fallthrough
			case strings.HasSuffix(fileName, ".m"): cpp = codemix.MixM(cpp)
			default: 
				fmt.Println("operate the file name : " + fileName);
				
				return -1
		}

		cppPath := *dstDir + fileName

		lastIdx := strings.LastIndex(cppPath, apartStr)
		dirName := cppPath[:lastIdx]
		if _, e := os.Stat(dirName); e != nil {

			os.MkdirAll(dirName, 0755)
		}
		ioutil.WriteFile(cppPath, cpp, os.ModePerm)

		if hasH {
			
			hPath := cppPath[:len(cppPath)-3] + "h"
			ioutil.WriteFile(hPath, h, os.ModePerm)
		}
		return 0
	})
}


func loadFile(dir *string, cbk func([]byte, []byte, string) int) {

	d, err := core.NewDirInfo(*dir, codemix.CodeFilter)
	if err != nil {
		fmt.Println("there is en error in loadFile : " + err.Error())
		return
	}

	apart := []byte{os.PathSeparator}
	for {

		name := d.NextFile()
		if name == "" {
			break
		}


		var h []byte
		
		switch {

			case strings.HasSuffix(name, ".cpp"):
				hName := name[:len(name)-4] + ".h"
				_, err := os.Stat(hName)
				if err != nil && !os.IsExist(err) {
					fmt.Println("no filename named : " + hName)			
					continue
				}
		
				h, err = ioutil.ReadFile(hName)
				if err != nil {
					fmt.Println("ioutil.ReadFile error : " + err.Error())
					continue
				}
		}

		cppM, err := ioutil.ReadFile(name)
		if err != nil {

			fmt.Println("ioutil.ReadFile error : " + err.Error())
			continue
		}
		t := []byte(name[len(*dir):])
		if bytes.HasPrefix(t, apart) {

			t = t[1:]
		}
		cbk(h, cppM, string(t))
	}
}
