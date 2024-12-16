import React, { useState, useEffect } from "react";
import './App.css';
import { API_ENDPOINTS } from "./apis";

function FileUploader() {
  const [file, setFile] = useState(null); // Stores the file
  const [previewUrl, setPreviewUrl] = useState(null); // Stores the preview URL

  const handleFileChange = (event) => {
    const selectedFile = event.target.files[0];
    if (selectedFile) {
      setFile(selectedFile);
      setPreviewUrl(URL.createObjectURL(selectedFile)); // Generate a preview URL
      console.log("File selected:", selectedFile);
    }
  };

  return (
    <div>
      <input
        type="file"
        accept="image/*"
        onChange={handleFileChange}
      />
      {previewUrl && (
        <div className="image-preview">
          <img src={previewUrl} alt="Selected File" style={{ maxWidth: "100%", maxHeight: "300px" }} />
        </div>
      )}
    </div>
  );
}

function submitPrint() {
  const submit = () => {
    console.log("Button clicked!");
  };

  return <button onClick={submit}>Print!</button>;
}

function Leaderboard() {
  const [leaderboard, setLeaderboard] = useState([]);

  useEffect(() => {
    async function fetchData() {
      try {
        const response = await fetch(API_ENDPOINTS.leaderboard);
        const data = await response.json();
        setLeaderboard(data);
      } catch (error) {
        console.error("Error fetching leaderboard data:", error);
      }
    }

    fetchData();
  }, []);

  return (
    <div>
      <h1>Leaderboard</h1>
      <table>
        <thead>
          <tr>
            <th>Rank</th>
            <th>Name</th>
            <th>Score</th>
            <th>Cost</th>
          </tr>
        </thead>
        <tbody>
          {leaderboard.length > 0 ? (
            leaderboard.map((item, index) => (
              <tr key={index}>
                <td>{item.rank}</td>
                <td>{item.name}</td>
                <td>{item.score}</td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan="3">No data available</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}




function App() {
  const [loadedImageUrl, setLoadedImageUrl] = useState(null);

  // Fetch image from API on page load
  useEffect(() => {
    async function fetchImage() {
      try {
        const response = await fetch(API_ENDPOINTS.getImage);
        if (!response.ok) {
          throw new Error("Failed to fetch image");
        }

        const data = await response.json();
        console.log("Image fetched on load:", data);
        setLoadedImageUrl(data.imageUrl); // Assuming API response includes image URL
      } catch (error) {
        console.error("Error fetching image:", error);
      }
    }

    fetchImage();
  }, []);

  return (
    <div className="App">
       <div className="last-uploaded-image">
        {loadedImageUrl ? (
          <img
            src={loadedImageUrl}
            alt="Loaded from API"
            style={{ maxWidth: "100%", maxHeight: "300px" }}
          />
        ) : (
          <p>Loading image...</p>
        )}
      </div>
      <div className="leaderboard-container">
        <Leaderboard />
      </div>
      <div className="file-uploader-container">
        <FileUploader />
      </div>
      <div className="submit-container">
        {submitPrint()}
      </div>
    </div>
  );
}

export default App;
