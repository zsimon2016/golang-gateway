package api

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/simon/server"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"runtime/debug"
	"strings"

	util "gitee.com/simon_git_code/goutil"
)

type Api struct {
	Service server.Service
}
type responseResult struct {
	Msg  interface{}            `json:"msg"`
	Data map[string]interface{} `json:"data"`
	// Data map[string]string `json:"data"`
	Code    int64 `json:"code"`
	Success bool  `json:"success"`
}

type ProxRequest struct {
	R       map[string]interface{}
	Service string
	Method  string
}
type Reply struct {
	Result map[string]interface{}
	// Detail map[string]string
}

func (api *Api) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", api.MainHandler)
	listen := ""
	if res, ok := api.Service.ConfMap["gatewayAddress"]; ok {
		listen = res.(string)
	} else {
		listen = "0.0.0.0:9092"
	}
	err := http.ListenAndServe(listen, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func (api *Api) Health(addr *string, fun func(w http.ResponseWriter, r *http.Request)) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", fun)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func (api *Api) MainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	output := make(map[string]interface{})
	output["code"] = 0
	output["success"] = false
	reply := new(Reply)
	var err error
	queryJson := map[string]interface{}{}
	queryStr := map[string]interface{}{}
	args := new(ProxRequest)

	defer func() {
		if err := recover(); err != nil {
			stack := debug.Stack()
			log.Println(err, string(stack))
			output["code"] = 54001
			output["msg"] = err
			jsonStr, _ := json.Marshal(output)
			fmt.Fprintf(w, string(jsonStr))
		}
		r.Body.Close()
	}()
	path := r.URL.Path
	raw := getRequestBody(r)
	body := ""
	bodyStr, uerr := url.QueryUnescape(raw)
	if uerr == nil {
		body = bodyStr
	} else {
		body = raw
	}

	newBody := ""
	if len(bodyStr) > 0 {
		newBody = bodyStr[1 : len(bodyStr)-1]
	}
	catstr := regexp.MustCompile(`{.*}`).FindAllString(newBody, -1)

	pathArr := strings.Split(path, "/")

	permission := pathArr[2]
	// uni universal ??????????????????
	// acc accredit  ????????????????????????
	// opt optional  ???????????????????????????????????????
	if permission != "uni" && permission != "acc" && permission != "opt" {
		panic("Illegal request")
	}
	// serviceName := pathArr[3]
	serviceName := (api.Service.ConfMap)["services_list"].(map[string]interface{})[pathArr[3]]

	if serviceName == nil {
		panic("Service does not exist")

	}

	method := pathArr[4]
	// fmt.Println("===>", (*api.Service.ConfMap)[permission].(map[string]interface{})[pathArr[3]], method)
	isVild := (api.Service.ConfMap)[permission].(map[string]interface{})[pathArr[3]].(map[string]interface{})[method]
	if isVild == nil {
		panic("Deny access")
	}
	// ??????????????????
	if isVild == "off" {
		output["code"] = 200
		output["success"] = "success"
		output["data"] = map[string]interface{}{}
		goto OUT
	}

	if len(catstr) != 0 && (method == "Acquire" || method == "CreateTask") {
		ret := catstr[0]
		subUrl := strings.Replace(ret, `"`, `'`, -1)
		body = strings.Replace(bodyStr, ret, subUrl, -1)
	}

	if path != "/" {
		service := ""
		if serviceName != nil {
			service = serviceName.(string)
			args.Service = util.UpFirst(service)
		}
		args.Method = util.UpFirst(method)
	}

	for k, v := range r.Form {
		queryStr[k] = v[0]
	}
	// fmt.Println("r.Form===>", r.Form)

	if body != "" {
		jsonErr := json.Unmarshal([]byte(body), &queryJson)

		if jsonErr == nil {
			for k, val := range queryJson {
				queryStr[k] = val
			}
		} else {
			output["msg"] = jsonErr
			// fmt.Println(jsonErr)
		}
	}
	args.R = queryStr

	args.R["body"] = body
	args.R["header"] = getRequestHeader(r)

	//??????????????????
	// if permission == "impower" {
	// 	token := args.R["header"].(map[string][]string)["Token"]
	// 	if token == nil {
	// 		output["msg"] = "Token????????????"
	// 		output["code"] = 401
	// 		goto OUT
	// 	}
	// 	user, e := api.Service.Redis.CacheGet("SYS_USER:" + token[0])
	// 	// fmt.Println(user,token[0])
	// 	if user == nil || e != nil {
	// 		if e != nil {
	// 			output["msg"] = "?????????????????????????????????"
	// 			fmt.Println("CacheGet Error: " + e.Error())
	// 			output["code"] = 500
	// 		} else {
	// 			output["msg"] = "?????????????????????"
	// 			output["code"] = 401
	// 		}
	// 		goto OUT
	// 	}
	// }

	if args.Service == "Actuator" {
		output["code"] = 0
		output["msg"] = "ok"
		output["success"] = true
		goto OUT
	}

	args.R["ip"] = util.RemoteIp(r)
	err = api.Service.ServicePool(args.Service).Call(ctx, args.Method, args, &reply)
	if err != nil {
		panic(fmt.Sprintf("failed to call: %v", err))
	}
	if reply.Result == nil || len(reply.Result) == 0 {
		output["msg"] = "Service busy, please try again later"
		output["code"] = 50400
	} else {
		output["code"] = reply.Result["code"]
		output["msg"] = reply.Result["msg"]
		if reply.Result["data"] != nil {
			output["data"] = reply.Result["data"]
		} else {
			output["data"] = make(map[string]interface{})
		}
		output["success"] = true
	}

OUT:
	jsonStr, _ := json.Marshal(output)
	w.Header().Set("content-type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")                                                            // ??????????????????????????????????????????url??????????????????url?????????cookie??????
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token") //header?????????
	w.Header().Add("Access-Control-Allow-Credentials", "true")                                                    //?????????true?????????ajax???????????????cookie??????
	w.Header().Add("Access-Control-Allow-Methods", "POST, GET")                                                   //??????????????????
	fmt.Fprint(w, string(jsonStr))
}

func getRequestHeader(r *http.Request) map[string][]string {
	return r.Header
}
func getRequestBody(r *http.Request) string {
	s, _ := ioutil.ReadAll(r.Body) //???  body ????????????????????? s
	if len(s) > 0 {
		return string(s) //?????????????????????????????????
	} else {
		return ""
	}

}
