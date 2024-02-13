import React, { useState } from 'react';

function App() {
  const [fileList, setFileList] = useState([]);
  const [selectedFile, setSelectedFile] = useState(null);

  const fetchList = async () => {
    try {
      const response = await fetch('http://localhost:9000/list');
      const data = await response.json();
      setFileList(data);
    } catch (error) {
      console.error('Error fetching uploaded files:', error);
    }
  };

  const extractVersion = (fileName) => {
    const parts = fileName.split('_');
    return parts.length === 2 ? parts[1].replace('.bin', '') : null;
  };

  const selectFile = (fileName) => {
    const version = extractVersion(fileName);
    setSelectedFile({ fileName, version });
  
    // Send the selected version to the backend
    if (version) {
      fetch(`http://localhost:9000/select/${version}`, {
        method: 'POST',
      })
        .then(response => response.text())
        .then(data => console.log(data))
        .catch(error => console.error('Error selecting file version:', error));
    }
  };

  return (
    <div>
      <form encType="multipart/form-data" action="http://localhost:9000/upload" method="post">
        <input type="file" name="myFile" />
        <input type="submit" value="Upload" />
      </form>

      <button onClick={fetchList}>Show List</button>

      <div style={{ marginTop: '20px' }}>
        {/* Add space below the "Show List" button */}
      </div>

      <div id="fileList" style={{ display: 'flex', flexDirection: 'column' }}>
        {/* The fetched list of files in the 'uploads' folder will be displayed here as buttons */}
        {fileList.map((file) => (
          <button
            key={file.Name}
            onClick={() => selectFile(file.Name)}
            style={{ marginBottom: '5px', color: selectedFile?.fileName === file.Name ? 'red' : 'black' }}
          >
            {file.Name} - Version: {extractVersion(file.Name)}
          </button>
        ))}
      </div>

      <div id="selectedFile">
        {/* Display information about the selected file */}
        {selectedFile && (
          <div>
            Selected File: {selectedFile.fileName} - Version: {selectedFile.version}
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
