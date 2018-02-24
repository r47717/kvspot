package main

import ( 
  "fmt"
  "os"
  "strings"
  "log"
  "net/http"
  "encoding/gob"
  "encoding/json"
)

const (
  VERSION = "1.0"
  DEVELOPER = "r47717"
  DATE = ""
  PORT = "8666"
  DUMP_FILE = "/var/kvspot/dump"
)

var kv = make(map[string]string)


func main() {
  http.HandleFunc("/", homeHandler)
  http.HandleFunc("/api/", apiHandler)
  loadDump()
  fmt.Println("Starting server on port " + PORT)
  log.Fatal(http.ListenAndServe(":" + PORT, nil))
}


func homeHandler(w http.ResponseWriter, r *http.Request) {
  about := fmt.Sprintf("<b>KV Spot, version: %s, Developer: %s, %s</b><br>", VERSION, DEVELOPER, DATE)
  fmt.Fprintf(w, about)
  fmt.Fprintf(w, "<b>Usage:</b><br>")
  fmt.Fprintf(w, "GET /api/put/key/value -> JSON data<br>")
  fmt.Fprintf(w, "GET /api/get/key -> JSON data<br>")
  fmt.Fprintf(w, "GET /api/clean -> JSON data")
}


func apiHandler(w http.ResponseWriter, r *http.Request) {
  params := r.URL.Path[len("/api/"):]
  paramsArr := strings.Split(params, "/")
  op := paramsArr[0]

  var success, data string

  switch op {
  case "get":
    key := paramsArr[1]
    success, data = get(key)
  case "put":
    key := paramsArr[1]
    val := paramsArr[2]
    success, data = put(key, val)
  case "clean":
    success, data = clean()  
  default:
    success = "false"
    data = fmt.Sprintf("invalid operation '%s'", op)
    return
  }

  response := map[string]string{
    "success": success,
    "data": data,
  }
  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write(js)
}


func get(key string) (string, string) {
  if val, ok := kv[key]; ok {
    return "true", val
  } 
  
  return "false", ""
}


func put(key string, value string) (string, string) {
  kv[key] = value
  dump()

  return "true", value
}


func clean() (string, string) {
  kv = make(map[string]string)
  dump()
  return "true", ""
}


/*
func logToFile(str string) {
  file, err := os.OpenFile(LOG_FILE, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0660);
  if err != nil {
    fmt.Println("Error: cannot open log file, " + err.Error())
    return
  }
  defer file.Close()

  _, err = file.WriteString(str)
  if err != nil {
    fmt.Println("Error: cannot write to log file, " + err.Error())
  }

}
*/


func dump() {
  file, err := os.Create(DUMP_FILE)
  if err != nil {
    log.Println("Error: cannot create dump file, " + err.Error());
    return
  }
  defer file.Close()

  encoder := gob.NewEncoder(file)

  if err := encoder.Encode(kv); err != nil {
    log.Println("Error: cannot encode data, " + err.Error());
  }
}


func loadDump() {
  file, err := os.Open(DUMP_FILE)
  if err != nil {
    log.Println("Error: cannot open dump, " + err.Error());
    return
  }
  defer file.Close()

  decoder := gob.NewDecoder(file)
  err = decoder.Decode(&kv)
  if err != nil {
    log.Println("Error: dump decoding failed:" + err.Error())
    return
  }
}

