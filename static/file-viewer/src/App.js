// import React, { useState } from 'react';

// function App() {
//   const [fileList, setFileList] = useState([]);
//   const [selectedDevice, setSelectedDevice] = useState(null);

//   // Fetch the list of files from the backend
//   const fetchList = async () => {
//     try {
//       const response = await fetch('http://localhost:9000/list', {
//         headers:{
//           'ngrok-skip-browser-warning': "dnwiind"
//         }
//       });
//       const data = await response.json();
//       setFileList(data);
//     } catch (error) {
//       console.error('Error fetching uploaded files:', error);
//     }
//   };

//   // Extract version from file name
//   const extractVersion = (fileName) => {
//     const parts = fileName.split('_');
//     return parts.length === 2 ? parts[1].replace('.bin', '') : null;
//   };

//   // Handle selecting a file and sending the version and IMEI to the backend
//   const selectFile = (fileName, version, imei) => {
//     fetch(`http://localhost:9000/select/${version}`, {
//       method: 'POST',
//       headers: {
//         'Content-Type': 'application/json',
//         'X-Device-IMEI': imei,  // Send IMEI as a header
//       }
//     })
//     .then(response => {
//       if (!response.ok) {
//         throw new Error('Network response was not ok');
//       }
//       return response.text();
//     })
//     .then(data => console.log(data))
//     .catch(error => console.error('Error selecting file version:', error));
//   };

//   // Handle device selection
//   const handleDeviceSelection = (event) => {
//     const imei = event.target.value;
//     setSelectedDevice({ imei });
//   };

//   // Handle file selection
//   const handleFileSelection = (fileName) => {
//     const version = extractVersion(fileName);
//     if (!selectedDevice) {
//       console.error('No device selected');
//       return;
//     }
//     const imei = selectedDevice.imei;
//     selectFile(fileName, version, imei);
//   };

//   return (
//     <div>
//       {/* Device selection dropdown */}
//       <div>
//         <label htmlFor="deviceSelect">Select Device:</label>
//         <select id="deviceSelect" onChange={handleDeviceSelection}>
//           <option value="">-- Select a Device --</option>
//           <option value="123456789012345">Device 1 (IMEI: 123456789012345)</option>
//           <option value="987654321098765">Device 2 (IMEI: 987654321098765)</option>
//           <option value="112233445566778">Device 3 (IMEI: 112233445566778)</option>
//         </select>
//       </div>

//       {/* Form to upload a file */}
//       <form encType="multipart/form-data" action="http://localhost:9000/upload" method="post">
//         <input type="file" name="myFile" />
//         <input type="submit" value="Upload" />
//       </form>

//       {/* Button to fetch the list of uploaded files */}
//       <button onClick={fetchList}>Show List</button>

//       <div style={{ marginTop: '20px' }}>
//         {/* Space added below the "Show List" button */}
//       </div>

//       {/* Display the list of uploaded files */}
//       <div id="fileList" style={{ display: 'flex', flexDirection: 'column' }}>
//         {fileList.map((file) => (
//           <div key={file.Name} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '5px' }}>
//             {/* Display the file name and version */}
//             <button
//               onClick={() => handleFileSelection(file.Name)}
//               style={{ textAlign: 'left', flexGrow: 1 }}
//             >
//               {file.Name} - Version: {extractVersion(file.Name)}
//             </button>
//             {/* Display the upload time on the right */}
//             <span style={{ marginLeft: '10px' }}>{file.UploadTime}</span>
//           </div>
//         ))}
//       </div>

//     </div>
//   );
// }

// export default App;


// import React, { useState } from "react";
// import mqtt from "mqtt";

// function App() {
//   const [fileList, setFileList] = useState([]);
//   const [selectedDevice, setSelectedDevice] = useState(null);
//   const [message, setMessage] = useState("");

//   // Connect to the MQTT broker
//   const client = mqtt.connect("ws://localhost:9001"); // WebSocket connection to the broker

//   // When the client connects
//   client.on("connect", () => {
//     console.log("Connected to MQTT broker");
//   });

//   // Fetch the list of files from the backend
//   const fetchList = async () => {
//     try {
//       const response = await fetch("http://localhost:9000/list", {
//         // headers: {
//         //   "ngrok-skip-browser-warning": "dnwiind",
//         // },
//       });
//       const data = await response.json();
//       setFileList(data);
//     } catch (error) {
//       console.error("Error fetching uploaded files:", error);
//     }
//   };

//   // Extract version from file name
//   const extractVersion = (fileName) => {
//     const parts = fileName.split("_");
//     return parts.length === 2 ? parts[1].replace(".bin", "") : null;
//   };

//   // Handle selecting a file and sending the version and IMEI to the backend
//   const selectFile = (fileName, version, imei) => {
//     fetch(`http://localhost:9000/select/${version}`, {
//       method: "POST",
//       headers: {
//         "Content-Type": "application/json",
//         "X-Device-IMEI": imei, // Send IMEI as a header
//       },
//     })
//       .then((response) => {
//         if (!response.ok) {
//           throw new Error("Network response was not ok");
//         }
//         return response.text();
//       })
//       .then((data) => console.log(data))
//       .catch((error) =>
//         console.error("Error selecting file version:", error)
//       );
//   };

//   // Handle device selection
//   const handleDeviceSelection = (event) => {
//     const imei = event.target.value;
//     setSelectedDevice({ imei });
//   };

//   // Handle file selection
//   const handleFileSelection = (fileName) => {
//     const version = extractVersion(fileName);
//     if (!selectedDevice) {
//       console.error("No device selected");
//       return;
//     }
//     const imei = selectedDevice.imei;
//     selectFile(fileName, version, imei);
//   };

//   // Function to handle sending the MQTT message to the 'switch/message' topic
//   const sendMessage = () => {
//     if (message) {
//       // Publish the message to the 'switch/message' topic
//       client.publish("switch/message", message, (error) => {
//         if (error) {
//           console.error("Error publishing message: ", error);
//         } else {
//           console.log("Message sent successfully:", message);
//         }
//       });

//       // Clear the input field after sending the message
//       setMessage("");
//     }
//   };

//   return (
//     <div>
//       {/* Device selection dropdown */}
//       <div>
//         <label htmlFor="deviceSelect">Select Device:</label>
//         <select id="deviceSelect" onChange={handleDeviceSelection}>
//           <option value="">-- Select a Device --</option>
//           <option value="123456789012345">Device 1 (IMEI: 123456789012345)</option>
//           <option value="987654321098765">Device 2 (IMEI: 987654321098765)</option>
//           <option value="112233445566778">Device 3 (IMEI: 112233445566778)</option>
//         </select>
//       </div>

//       {/* Form to upload a file */}
//       <form
//         encType="multipart/form-data"
//         action="http://localhost:9000/upload"
//         method="post"
//       >
//         <input type="file" name="myFile" />
//         <input type="submit" value="Upload" />
//       </form>

//       {/* Button to fetch the list of uploaded files */}
//       <button onClick={fetchList}>Show List</button>

//       <div style={{ marginTop: "20px" }}>
//         {/* Space added below the "Show List" button */}
//       </div>

//       {/* Display the list of uploaded files */}
//       <div id="fileList" style={{ display: "flex", flexDirection: "column" }}>
//         {fileList.map((file) => (
//           <div
//             key={file.Name}
//             style={{
//               display: "flex",
//               alignItems: "center",
//               justifyContent: "space-between",
//               marginBottom: "5px",
//             }}
//           >
//             {/* Display the file name and version */}
//             <button
//               onClick={() => handleFileSelection(file.Name)}
//               style={{ textAlign: "left", flexGrow: 1 }}
//             >
//               {file.Name} - Version: {extractVersion(file.Name)}
//             </button>
//             {/* Display the upload time on the right */}
//             <span style={{ marginLeft: "10px" }}>{file.UploadTime}</span>
//           </div>
//         ))}
//       </div>

//       {/* MQTT Publisher Section */}
//       <div style={{ marginTop: "20px" }}>
//         <h2>MQTT Publisher</h2>
//         <input
//           type="text"
//           value={message}
//           onChange={(e) => setMessage(e.target.value)}
//           placeholder="Enter message"
//         />
//         <button onClick={sendMessage}>Send</button>
//       </div>
//     </div>
//   );
// }

// export default App;

import React, { useState } from "react";
import mqtt from "mqtt";

function App() {
  const [fileList, setFileList] = useState([]);
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [message, setMessage] = useState("");

  // Connect to the MQTT broker
  const client = mqtt.connect("ws://localhost:9001"); // WebSocket connection to the broker

  client.on("connect", () => {
    console.log("Connected to MQTT broker");
  });

  // Fetch the list of files from the backend
  const fetchList = async () => {
    try {
      const response = await fetch("http://localhost:9000/list");
      const data = await response.json();
      setFileList(data);
    } catch (error) {
      console.error("Error fetching uploaded files:", error);
    }
  };

  // Extract version from file name
  const extractVersion = (fileName) => {
    const parts = fileName.split("_");
    return parts.length === 2 ? parts[1].replace(".bin", "") : null;
  };

  // Handle selecting a file and sending the version and IMEI to the backend
  const selectFile = (fileName, version, imei) => {
    fetch(`http://localhost:9000/select/${version}/${imei}`, {
      method: "POST",
    })
    .then((response) => {
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.text();
    })
    .then((data) => console.log(data))
    .catch((error) => console.error("Error selecting file version:", error));
  };

  // Handle device selection
  const handleDeviceSelection = (event) => {
    const imei = event.target.value;
    setSelectedDevice({ imei });
  };

  // Handle file selection
  const handleFileSelection = (fileName) => {
    const version = extractVersion(fileName);
    if (!selectedDevice) {
      console.error("No device selected");
      return;
    }
    const imei = selectedDevice.imei;
    selectFile(fileName, version, imei);
  };

  // Function to handle sending the MQTT message
  const sendMessage = () => {
    if (message) {
      client.publish("switch/message", message, (error) => {
        if (error) {
          console.error("Error publishing message:", error);
        } else {
          console.log("Message sent successfully:", message);
        }
      });
      setMessage("");
    }
  };

  return (
    <div>
      <div>
        <label htmlFor="deviceSelect">Select Device:</label>
        <select id="deviceSelect" onChange={handleDeviceSelection}>
          <option value="">-- Select a Device --</option>
          <option value="123456789012345">Device 1 (IMEI: 123456789012345)</option>
          <option value="987654321098765">Device 2 (IMEI: 987654321098765)</option>
          <option value="112233445566778">Device 3 (IMEI: 112233445566778)</option>
        </select>
      </div>

      <form encType="multipart/form-data" action="http://localhost:9000/upload" method="post">
        <input type="file" name="myFile" />
        <input type="submit" value="Upload" />
      </form>

      <button onClick={fetchList}>Show List</button>

      <div style={{ marginTop: "20px" }}></div>

      <div id="fileList" style={{ display: "flex", flexDirection: "column" }}>
        {fileList.map((file) => (
          <div
            key={file.Name}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              marginBottom: "5px"
            }}
          >
            <button
              onClick={() => handleFileSelection(file.Name)}
              style={{ textAlign: "left", flexGrow: 1 }}
            >
              {file.Name} - Version: {extractVersion(file.Name)}
            </button>
            <span style={{ marginLeft: "10px" }}>{file.UploadTime}</span>
          </div>
        ))}
      </div>

      <div style={{ marginTop: "20px" }}>
        <h2>MQTT Publisher</h2>
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Enter message"
        />
        <button onClick={sendMessage}>Send</button>
      </div>
    </div>
  );
}

export default App;
