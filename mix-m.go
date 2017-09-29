package codemix

import (
	"strconv"
	"fmt"
	"time"
	"math/rand"
	"bytes"
)

// OCFuncInfo 用于维护oc函数信息
type OCFuncInfo struct {

	params []*OCParam
}

// OCParam 用于维护oc函数中参数信息
type OCParam struct {
	
	lName string
	rName string

	t string
}

// MixM 返回处理后的.m文件内容
func MixM(buf []byte) (rlt []byte) {

	tmpBuf := cleanBuf(buf)

	endIdx := bytes.LastIndex(tmpBuf, []byte("@end"))

	if endIdx == -1 {

		fmt.Println("cannot find end tag")
		return buf
	}

	// beginIdx := bytes.Index(buf, []byte("@implementation "))
	// beginIdx += bytes.IndexByte(buf[beginIdx:], '\n') + 1
	// beginBuf = buf[:beginIdx]
	// rlt = append(rlt, buf[:beginIdx]...)
	// buf = buf[beginIdx:endIdx]
	ts := []byte("@implementation")
	cnIdx := bytes.Index(tmpBuf, ts)
	cnBuf := tmpBuf[cnIdx+len(ts):]
	tmpIdx := bytes.IndexByte(cnBuf, '\n')
	cnBuf = bytes.TrimFunc(cnBuf[:tmpIdx], trimFiled)
	tmpIdx = bytes.IndexFunc(cnBuf, func(re rune) bool {

		rlt := false
		switch re {
			case ' ': fallthrough
			case ':': rlt = true
		}
		return rlt
	})
	cnBuf = cnBuf[:tmpIdx]


	
	rand.Seed(time.Now().UnixNano())
	num := 20 + rand.Intn(80)
	fi := make([]*OCFuncInfo, num)
	randOCFunc(fi)


	// var lastFuncIdx int
	lastFuncIdx := endIdx
	L: for {
		lastFuncIdx = bytes.LastIndexFunc(tmpBuf[:lastFuncIdx], func(re rune) bool {
			
					var rlt bool
					switch re {
			
						case '-': fallthrough
						case '+': rlt = true
						default: rlt = false
					}
					return rlt
				})
		// if buf[lastFuncIdx-1]
		switch {
			case lastFuncIdx < 1: fallthrough
			case tmpBuf[lastFuncIdx-1] == '\n': break L
		}
	}

	// rlt = make([]byte, endIdx, len(buf))
	// copy(rlt, buf[:endIdx])
	
	// rlt = append(rlt, toOCCallType(cnBuf, fi)...)
	fmd := toOCCallType(cnBuf, fi)
	for _, v := range fi[1:] {

		fmd = append(fmd, toOCBaseType(v)...)
	}
	
	// rlt = append(rlt, endBuf...)
	bufLen, fmdLen := len(buf), len(fmd)
	switch lastFuncIdx {
		case -1:
			rlt = make([]byte, bufLen + fmdLen)
			copy(rlt, buf[:endIdx])
			copy(rlt[endIdx:], fmd)
			copy(rlt[endIdx+fmdLen:], buf[endIdx:])
		default:

			r := nextOCFuncSite(tmpBuf[lastFuncIdx:])
			funcSlice := toOCAddCallFuncBytes(buf[lastFuncIdx:lastFuncIdx+r], cnBuf, fi[0])
			
			rlt = make([]byte, lastFuncIdx, bufLen)
			copy(rlt, buf[:lastFuncIdx])
			rlt = append(rlt, funcSlice...)
			rlt = append(rlt, fmd...)
			rlt = append(rlt, buf[lastFuncIdx+r:]...)
	}
	return
}


func toOCAddCallFuncBytes(buf []byte, cn []byte, of *OCFuncInfo) (rlt []byte) {

	idx := bytes.LastIndexByte(buf, '}')



	pre := []byte("\n\n\n\tint _mix_a = 7;\n\tint _mix_b = 4;\n\tbool _mix_sd = false;\n\tfor(int c=0; c<_mix_b*0.5; c++) {\n\t\t_mix_b += 1;\n\t}\n\t_mix_sd = _mix_a > _mix_b;\n\tif (_mix_sd)\n\t")
	cb := append(pre, toOCFuncCallBytes(cn, of)...)
	bufLen, cbLen := len(buf), len(cb)
	rltLen := bufLen + cbLen + 2
	rlt = make([]byte, rltLen)
	
	
	copy(rlt, buf[:idx])
	copy(rlt[rltLen-bufLen+idx:], buf[idx:])
	rlt[idx] = '\n'
	idx++
	copy(rlt[idx:], cb)
	rlt[idx+cbLen] = '\n'
	return
}


func toOCFuncCallBytes(cn []byte, of *OCFuncInfo) (rlt []byte) {

	cl := len(cn)
	rlt = make([]byte, cl+2, cl + 10)

	// rlt[0] = '['
	copy(rlt, []byte{'\t', '['})
	copy(rlt[2:], cn)
	
	// rlt = append(rlt, )
	for _, v := range of.params {

		pm := randOCParamValue(v)
		rlt = append(rlt, ' ')
		rlt = append(rlt, pm...)
	}
	rlt = append(rlt, ']', ';', '\n')
	return
}

func toOCBaseType(of *OCFuncInfo) []byte {

	n := toOCFullName(of)

	for _, item := range of.params {

		var tag byte
		switch item.t {

			case "int": tag = 'd'
			case "float": tag = 'f'
			case "NSString *": tag = '@'
		}

		log := []byte("\tNSLog(@\"%@")
		switch item.rName {
			case "":
				log = append(log, '"', ',', ' ', '@', '"')
				log = append(log, []byte(item.lName)...)
				log = append(log, '"', ')', ';', '\n')
			default:
				rb := []byte(item.rName)
				log = append(log, '=', '%', tag, '"', ',', ' ', '@', '"')
				log = append(log, rb...)
				log = append(log, '"', ',', ' ')
				log = append(log, rb...)
				log = append(log, ')', ';', '\n')
		}
		n = append(n, log...)
	}
	
	n = append(n, '}', '\n')
	return n
}

func toOCCallType(cn []byte, list []*OCFuncInfo) (rlt []byte) {

	// info := list[0]

	n := toOCFullName(list[0])
	
	for _, info := range list[1:] {


		b := toOCFuncCallBytes(cn, info)
		n = append(n, b...)
	}

	rlt = append(n, '}', '\n')
	return
}

func toOCFullName(of *OCFuncInfo) (rlt []byte) {


	rlt = []byte{'\n', '+', ' ', '(', 'v', 'o', 'i', 'd', ')'}
	for _, v := range of.params {

		rlt = append(rlt, []byte(v.lName)...)
		if v.rName != "" {

			rlt = append(rlt, ':', '(')
			rlt = append(rlt, []byte(v.t)...)
			rlt = append(rlt, ')')
			rlt = append(rlt, []byte(v.rName)...)
		}
		rlt = append(rlt, ' ')
	}
	rlt = append(rlt, '{', '\n')
	return
}


func randOCParamValue(op *OCParam) (pm []byte) {

	pm = make([]byte, len(op.lName))
	copy(pm, []byte(op.lName))

	var rlt []byte
	switch op.t {

		case "NSString *":
			l := len(op.t)
			rlt = make([]byte, 3 + l)
			copy(rlt, []byte{'@', '"'})
			copy(rlt[2:], []byte(op.t))
			rlt[l+2] = '"'
		case "int": rlt = []byte(strconv.Itoa(rand.Intn(1024)))
		case "float": rlt = []byte(strconv.FormatFloat(rand.Float64()*666, 'f', 2, 64))
		default: return
	}

	pm = append(pm, ':')
	pm = append(pm, rlt...)

	return
}


func randOCFunc(list []*OCFuncInfo) {

	for i := range list {

		oi := new(OCFuncInfo)
		oi.params = randOCParam(list[:i])

		list[i] = oi
	}	
}





func randOCParam(ocl []*OCFuncInfo) (list []*OCParam) {


		
	sameName := func(l []*OCFuncInfo, name string) (rlt bool) {
		
				rlt = false
				for _, v := range l {
		
					switch {
						case v == nil: fallthrough
						case v.params == nil: fallthrough
						case len(v.params) == 0: fallthrough
						case v.params[0].lName != name: continue
					}
					rlt = true
				}
				return
			}

	num := rand.Intn(4)
	nameList := make([]string, 0, num*2)
	listNum := num
	if num == 0 {
		listNum = 1
	}
	list = make([]*OCParam, listNum)

	param := new(OCParam)
	fn := randParamName(nameList)
	for sameName(ocl, fn) {

		fn = randParamName(nameList)
	}
	// nameList = append(nameList, fn)
	
	*param = OCParam{

		lName : fn,
	}
	list[0] = param
	
	if num > 0 {
		
		nameList = append(nameList, fn)
		rName := randParamName(nameList)
		nameList = append(nameList, rName)
		param.rName = rName
		param.t = randOCType()
		
		for i:=1; i<num; i++ {

			name := randParamName(nameList)
			nameList = append(nameList, name)
			p := randParamName(nameList)
			nameList = append(nameList, p)

			param := new(OCParam)
			*param = OCParam{
				lName : name,
				rName : p,
				t : randOCType(),
			}

			list[i] = param
		}
	}
	return
}


func randOCType() (t string) {
	
	switch rand.Intn(3) {
		case 0: t = "NSString *"
		case 1: t = "int"
		case 2: t = "float"
		// case 3: t = "void"
		default: t = "void"
	}
	return
}


// func nextOCFuncSite(buf []byte) (lIdx int, rIdx int) {

// 	bufSlice := buf
// 	L: for {
		

// 		l := bytes.IndexFunc(bufSlice, ocFuncFlag)
// 		switch {
// 			case l == -1: 
// 				return -1, -1
// 			case l == 0:
// 			case bufSlice[l-1] != '\n': 
// 				bufSlice = bufSlice[l+1:]
// 				continue L
// 		}

// 		bufSlice = bufSlice[l:]
// 		lIdx = len(buf) - len(bufSlice)
		
		
// 		l = bytes.IndexByte(bufSlice, '{')
// 		if l == -1 {
// 			return -1, -1
// 		}

// 		bufSlice = bufSlice[l:]
// 		num := 0
// 		var line []byte
// 		for {
			
// 			if len(bufSlice) == 0 {
// 				return -1, -1
// 			}

// 			idx = bytes.IndexByte(bufSlice, '\n')
// 			if idx == -1 {
// 				line = bufSlice
// 				bufSlice = []byte{}
// 			} else {
// 				line = bufSlice[:idx]
// 				bufSlice = bufSlice[idx+1:]
// 			}
// 			num = braceNum(line, num)

// 			if num == 0 {

// 				rIdx = len(buf) - len(bufSlice)
// 				break L
// 			}
// 		}
// 	}
// 	return
// }


func ocFuncFlag(r rune) (flag bool) {

	switch r {
		case '+': fallthrough
		case '-': flag = true
		default: flag = false
	}
	return
}



func nextOCFuncSite(oc []byte) int {
	
	// fmt.Println("nextFuncSite :" + string(cpp))
	idx := bytes.IndexByte(oc, '{')
	if idx == -1 {
		return -1
	}

	i := nextBraceIndex(oc[idx:])
	if i == -1 {


		fmt.Println("找不到oc函数!!!")
		return -1
	}

	return idx + i
}

