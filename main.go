package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type FileInfo struct {
	Name string
	UploadTime string `json:"uploadTime"`
}

// VersionInfo stores the version selected for each device (identified by IMEI)
type VersionInfo struct {
	Version string
	IMEI    string
}

var mu sync.Mutex
var selectedVersion string
var deviceVersions = make(map[string]VersionInfo) // Map to store the selected version for each device

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error Retrieving the File", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header.Get("Content-Type"))

	// Record the current date and time
	uploadTime := time.Now().Format("2006-01-02 15:04:05")

	tempFile, err := os.Create(filepath.Join("uploads", handler.Filename))
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// Store the upload time in a separate file with the same name and a .time extension
	timeFile, err := os.Create(filepath.Join("uploads", handler.Filename+".time"))
	if err != nil {
		http.Error(w, "Error creating time file", http.StatusInternalServerError)
		return
	}
	defer timeFile.Close()

	timeFile.WriteString(uploadTime)

	fmt.Fprintf(w, "Successfully Uploaded File: %s\n", handler.Filename)
}

// var lastServedVersion string

func findFileByVersion(requestedVersion string) string {
	files, err := filepath.Glob(filepath.Join("uploads", "esp32_"+requestedVersion+".bin"))
	if err != nil {
		return ""
	}

	if len(files) > 0 {
		return filepath.Base(files[0])
	}
	return ""
}

func serveFileWithProgress(w http.ResponseWriter, r *http.Request) {
    fmt.Println("update begun.....")
    vars := mux.Vars(r)

    requestedVersion, ok := vars["version"]
    if !ok {
        http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
        return
    }

    imei, ok := vars["imei"]
    if !ok || imei == "" {
        http.Error(w, "IMEI not provided in the URL", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    // Retrieve the version information for the given IMEI
    versionInfo, exists := deviceVersions[imei]
    if !exists {
        http.Error(w, "No version selected for this IMEI", http.StatusBadRequest)
        return
    }

    if versionInfo.Version != requestedVersion {
        filename := findFileByVersion(versionInfo.Version)
        if filename == "" {
            http.Error(w, "Requested version not found", http.StatusNotFound)
            return
        }

        filePath := filepath.Join("uploads", filename)
        file, err := os.Open(filePath)
        if err != nil {
            http.Error(w, "Error opening file", http.StatusInternalServerError)
            return
        }
        defer file.Close()

        fi, err := file.Stat()
        if err != nil {
            http.Error(w, "Error getting file information", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Disposition", "attachment; filename="+filename)
        w.Header().Set("Content-Type", "application/octet-stream")
        w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

        const bufferSize = 8192
        buffer := make([]byte, bufferSize)

        totalBytesRead := int64(0)

        for {
            n, err := file.Read(buffer)
            if err == io.EOF {
                break
            } else if err != nil {
                http.Error(w, "Error reading file", http.StatusInternalServerError)
                return
            }

            _, err = w.Write(buffer[:n])
            if err != nil {
                return
            }

            totalBytesRead += int64(n)

            if flusher, ok := w.(http.Flusher); ok {
                flusher.Flush()
            }
        }
        log.Printf("Download of %s completed: %d bytes\n", filename, totalBytesRead)
    } else {
        errorMessage := "Firmware is up to date. No update required."
        http.Error(w, errorMessage, http.StatusBadRequest)
        fmt.Fprintf(w, "%s", errorMessage)
        return
    }
}



// func serveFileWithProgress(w http.ResponseWriter, r *http.Request) {
//     fmt.Println("update begun.....")
//     vars := mux.Vars(r)

//     requestedVersion, ok := vars["version"]
//     if !ok {
//         http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
//         return
//     }

//     mu.Lock()
//     defer mu.Unlock()

//     if selectedVersion == "" {
//         http.Error(w, "No version selected", http.StatusBadRequest)
//         return
//     }

//     if selectedVersion != requestedVersion {
//         filename := findFileByVersion(selectedVersion)
//         if filename == "" {
//             http.Error(w, "Requested version not found", http.StatusNotFound)
//             return
//         }

//         filePath := filepath.Join("uploads", filename)
//         file, err := os.Open(filePath)
//         if err != nil {
//             http.Error(w, "Error opening file", http.StatusInternalServerError)
//             return
//         }
//         defer file.Close()

//         fi, err := file.Stat()
//         if err != nil {
//             http.Error(w, "Error getting file information", http.StatusInternalServerError)
//             return
//         }

//         w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

//         const bufferSize = 8192
//         buffer := make([]byte, bufferSize)
//         totalBytesRead := int64(0)
//         totalSize := fi.Size()

//         for {
//             n, err := file.Read(buffer)
//             if err == io.EOF {
//                 break
//             } else if err != nil {
//                 http.Error(w, "Error reading file", http.StatusInternalServerError)
//                 return
//             }

//             _, err = w.Write(buffer[:n])
//             if err != nil {
//                 return
//             }

//             totalBytesRead += int64(n)
//             log.Printf("Progress: %d / %d bytes (%.2f%%)\n", totalBytesRead, totalSize, (float64(totalBytesRead)/float64(totalSize))*100)

//             if flusher, ok := w.(http.Flusher); ok {
//                 flusher.Flush()
//             }
//         }
//         log.Printf("Download of %s completed: %d bytes\n", filename, totalBytesRead)
//     } else {
//         errorMessage := "Firmware is up to date. No update required."
//         http.Error(w, errorMessage, http.StatusBadRequest)
//         fmt.Fprintf(w, "%s", errorMessage)
//     }
// }


func listFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Showing the contents...")
	files, err := getUploadedFiles("uploads")
	if err != nil {
		http.Error(w, "Error listing files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func selectVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version, ok := vars["version"]
	if !ok {
		http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
		return
	}

	// Extract the IMEI from the request header
	imei := r.Header.Get("X-Device-IMEI")
	if imei == "" {
		http.Error(w, "IMEI not provided", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Store the selected version and IMEI in the map
	deviceVersions[imei] = VersionInfo{
		Version: version,
		IMEI:    imei,
	}

	fmt.Printf("Selected version for IMEI %s: %s\n", imei, version)

	fmt.Fprint(w, "Version selected successfully")
}
func getUploadedFiles(uploadDir string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(info.Name(), ".time") {
			// Read the corresponding .time file to get the upload time
			timeFilePath := filepath.Join(uploadDir, info.Name()+".time")
			uploadTime := "Unknown"
			if timeBytes, err := os.ReadFile(timeFilePath); err == nil {
				uploadTime = string(timeBytes)
			}

			// Populate the FileInfo struct with both the name and upload time
			files = append(files, FileInfo{
				Name:       info.Name(),
				UploadTime: uploadTime,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
func main() {
	r := mux.NewRouter()
    r.HandleFunc("/upload", uploadFile).Methods("POST")
    r.HandleFunc("/list", listFiles).Methods("GET")
    r.HandleFunc("/select/{version:[0-9.]+}", selectVersion).Methods("POST")
    r.HandleFunc("/{filename}/{version:[0-9.]+}/{imei}", serveFileWithProgress).Methods("GET")


    // CORS configuration
    corsOptions := handlers.CORS(
        handlers.AllowedOrigins([]string{"http://localhost:3000"}),
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        handlers.AllowedHeaders([]string{"Content-Type", "X-Device-IMEI"}),
    )

    // Apply the CORS middleware
    http.ListenAndServe(":9000", corsOptions(r))

    log.Println("Listening on 9000...")
}

//send the flag with the imei,timestamp
//send timestamp of when it was uploaded
//when making the request include the imei




