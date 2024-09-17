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
	FileName string
}

var mu sync.Mutex

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

	// Record the current date and time in ISO 8601 format with milliseconds
	uploadTime := time.Now().Format("2006-01-02T15:04:05.000Z")

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

// func findFileByVersion(requestedVersion string) string {
// 	files, err := filepath.Glob(filepath.Join("uploads", "esp32_"+requestedVersion+".bin"))
// 	// files, err := filepath.Glob(filepath.Join("uploads", "*"+requestedVersion+"*.bin"))
// 	if err != nil {
// 		return ""
// 	}

// 	if len(files) > 0 {
// 		return filepath.Base(files[0])
// 	}
// 	return ""
// }
func findFileByVersion(fileNameWithoutVersion, requestedVersion string) string {
	// Use a wildcard to match the file with any name but containing the specific version
	pattern := filepath.Join("uploads", fileNameWithoutVersion+"_"+requestedVersion+"*.bin")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return ""
	}

	// If matching files are found, return the first one
	if len(files) > 0 {
		return filepath.Base(files[0])
	}

	// Return an empty string if no matching file is found
	return ""
}

// func serveFileWithProgress(w http.ResponseWriter, r *http.Request) {
//     fmt.Println("update begun.....")
//     vars := mux.Vars(r)

//     requestedVersion, ok := vars["version"]
//     if !ok {
//         http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
//         return
//     }

//     imei, ok := vars["imei"]
//     if !ok || imei == "" {
//         http.Error(w, "IMEI not provided in the URL", http.StatusBadRequest)
//         return
//     }

//     mu.Lock()
//     defer mu.Unlock()

//     // Retrieve the version information for the given IMEI
//     versionInfo, exists := deviceVersions[imei]
//     if !exists {
//         http.Error(w, "No version selected for this IMEI", http.StatusBadRequest)
//         return
//     }

//     if versionInfo.Version != requestedVersion {
//         filename := findFileByVersion(versionInfo.Version)
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

//         w.Header().Set("Content-Disposition", "attachment; filename="+filename)
//         w.Header().Set("Content-Type", "application/octet-stream")
//         w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

//         const bufferSize = 8192
//         buffer := make([]byte, bufferSize)

//         totalBytesRead := int64(0)

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

//             if flusher, ok := w.(http.Flusher); ok {
//                 flusher.Flush()
//             }
//         }
//         log.Printf("Download of %s completed: %d bytes\n", filename, totalBytesRead)
//     } else {
//         errorMessage := "Firmware is up to date. No update required."
//         http.Error(w, errorMessage, http.StatusBadRequest)
//         fmt.Fprintf(w, "%s", errorMessage)
//         return
//     }
// }
func serveFileWithProgress(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Update begun...")

    // Extract variables from the URL
    vars := mux.Vars(r)

    // Extract the requested version from the URL
    requestedVersion, ok := vars["version"]
    if !ok {
        http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
        return
    }

    // Extract the IMEI from the URL
    imei, ok := vars["imei"]
    if !ok || imei == "" {
        http.Error(w, "IMEI not provided in the URL", http.StatusBadRequest)
        return
    }

    // Extract the base file name from the URL
    baseFileName, ok := vars["filename"]
    if !ok || baseFileName == "" {
        http.Error(w, "Base file name not provided in the URL", http.StatusBadRequest)
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
	
	

    // Check if the firmware is already up to date
    if versionInfo.Version == requestedVersion {
        errorMessage := "Firmware is up to date. No update required."
        http.Error(w, errorMessage, http.StatusBadRequest)
        fmt.Fprintf(w, "%s", errorMessage)
        return
    }

	//check if the file names match
	if versionInfo.FileName != baseFileName{
		errorMessage1 := "FileNames donot match."
		http.Error(w, errorMessage1, http.StatusBadRequest)
		fmt.Fprintf(w, "%s", errorMessage1)
		return
	}
	
    // Use the updated findFileByVersion function to search for the file using the baseFileName and requestedVersion
    filename := findFileByVersion(baseFileName, versionInfo.Version)
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

    // Get the file size information
    fi, err := file.Stat()
    if err != nil {
        http.Error(w, "Error getting file information", http.StatusInternalServerError)
        return
    }

    // Set headers for the file download
    w.Header().Set("Content-Disposition", "attachment; filename="+filename)
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

    const bufferSize = 8192
    buffer := make([]byte, bufferSize)
    totalBytesRead := int64(0)

    // Stream the file in chunks
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

        // Flush the buffer to ensure data is sent to the client immediately
        if flusher, ok := w.(http.Flusher); ok {
            flusher.Flush()
        }
    }

    // Log the completion of the download
    log.Printf("Download of %s completed: %d bytes\n", filename, totalBytesRead)
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

// func selectVersion(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	version, ok := vars["version"]
// 	if !ok {
// 		http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
// 		return
// 	}

// 	// Extract the IMEI from the request header
// 	imei := r.Header.Get("X-Device-IMEI")
// 	if imei == "" {
// 		http.Error(w, "IMEI not provided", http.StatusBadRequest)
// 		return
// 	}

// 	mu.Lock()
// 	defer mu.Unlock()

// 	// Store the selected version and IMEI in the map
// 	deviceVersions[imei] = VersionInfo{
// 		Version: version,
// 		IMEI:    imei,
// 	}

// 	fmt.Printf("Selected version for IMEI %s: %s\n", imei, version)

// 	fmt.Fprint(w, "Version selected successfully")
// }
func selectVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Extract version and IMEI from the URL parameters
	version, ok := vars["version"]
	if !ok {
		http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
		return
	}
	imei, ok := vars["imei"]
	if !ok {
		http.Error(w, "IMEI not provided in the URL", http.StatusBadRequest)
		return
	}
	fileName, ok := vars["filename"]
	if !ok {
		http.Error(w, "Filename not provided in the URL", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Store the selected version and IMEI in the map
	deviceVersions[imei] = VersionInfo{
		Version: version,
		IMEI:    imei,
		FileName: fileName,
	}

	fmt.Printf("Selected file %s of version %s with IMEI %s\n", fileName,version, imei)
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
func deleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	if filename == "" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("uploads", filename)
	err := os.Remove(filePath)
	if err != nil {
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	// Optionally, delete the associated .time file
	timeFilePath := filePath + ".time"
	os.Remove(timeFilePath) // We ignore the error here as it is not critical

	fmt.Fprintf(w, "File %s deleted successfully", filename)
}

func main() {

	// var broker = "localhost"
	// var port = 1883
	// fmt.Println("Starting application")

	// Initialize MQTT service
	// mqttService, err := NewMQTTClient(port, broker, "", "")
	// if err != nil {
	// 	panic(err)
	// }

	// topic := "switch/message"
	// // Channel for receiving subscribed messages
	// respDataChan := make(chan []byte)

	// // Start the subscription goroutine
	// go func() {
	// 	if err := mqttService.Subscribe(topic, respDataChan); err != nil {
	// 		fmt.Printf("Error subscribing: %v\n", err)
	// 	}
	// }()

	// go func() {
	// 	for {
	// 		var input string
	// 		fmt.Print("Enter message to publish (or type 'exit' to quit): ")
	// 		fmt.Scanf("%s", &input)

	// 		// Exit the loop if the user types 'exit'
	// 		if input == "exit" {
	// 			close(respDataChan)
	// 			break
	// 		}

	// 		// Publish the user's message to the MQTT topic
	// 		message, _ := json.Marshal(&publishMessage{Message: input})
	// 		if err := mqttService.Publish(topic, message); err != nil {
	// 			fmt.Println("Error publishing:", err)
	// 		}
	// 		time.Sleep(time.Second)
	// 	}
	// }()
	r := mux.NewRouter()
	r.HandleFunc("/upload", uploadFile).Methods("POST")
	r.HandleFunc("/list", listFiles).Methods("GET")
	r.HandleFunc("/select/{version:[0-9.]+}/{imei:[0-9]+}/{filename}", selectVersion).Methods("POST")

	r.HandleFunc("/{filename}/{version:[0-9.]+}/{imei}", serveFileWithProgress).Methods("GET")
	r.HandleFunc("/delete/{filename}", deleteFile).Methods("DELETE")


    // CORS configuration
    corsOptions := handlers.CORS(
        handlers.AllowedOrigins([]string{"*"}),
        handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "ngrok-skip-browser-warning","X-Device-IMEI"}), // Include your custom header here
        // handlers.AllowCredentials(),
    )

    // Apply the CORS middleware
     http.ListenAndServe(":9000", corsOptions(r))
	log.Println("Listening on 9000...")
	// for msg := range respDataChan {
	// 	fmt.Printf("Received MQTT message: %s\n", string(msg))
	// }
	// close(respDataChan)
}

//send the flag with the imei,timestamp
//send timestamp of when it was uploaded
//when making the request include the imei




