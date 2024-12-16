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
  const submit = async () => {
    console.log("Button clicked!");

    try {
      // Get the user's IP address using an external API
      const ipResponse = await fetch("https://api.ipify.org?format=json");
      const ipData = await ipResponse.json();
      const userIP = ipData.ip; // The user's IP address

      console.log("User IP:", userIP);

      // Call the sendPrint endpoint
      const sendPrintResponse = await fetch(API_ENDPOINTS.sendPrint, {
        method: "POST",
        headers: {
          "X-API-Key": process.env.REACT_APP_API_KEY,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ message: "Print job initiated" }), // Adjust payload as needed
      });

      console.log(sendPrintResponse);

      if (!sendPrintResponse.ok) {
        throw new Error(`sendPrint failed: ${sendPrintResponse.statusText}`);
      }

      console.log("sendPrint succeeded");

      const insertLeaderboardResponse = await fetch(API_ENDPOINTS.insertLeaderboard, {
        method: "POST",
        headers: {
          "X-API-Key": process.env.REACT_APP_API_KEY,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: userIP }), 
      });

      console.log(process.env.REACT_APP_API_KEY);

      if (!insertLeaderboardResponse.ok) {
        throw new Error(`insertLeaderboard failed: ${insertLeaderboardResponse.statusText}`);
      }

      console.log("insertLeaderboard succeeded");
    } catch (error) {
      console.error("Error submitting print and leaderboard data:", error);
    }
  };

  return <button onClick={submit}>Print!</button>;
}


function Leaderboard() {
  const [leaderboard, setLeaderboard] = useState(null);  // Initialize with null instead of an empty array

  useEffect(() => {
    async function fetchData() {
      try {
        const response = await fetch(API_ENDPOINTS.leaderboardRead, {
          headers: {
            "X-API-Key": process.env.REACT_APP_API_KEY,
            "Content-Type": "application/json",
          },
        });
        const data = await response.json();
        setLeaderboard(data);  // Set the leaderboard data
      } catch (error) {
        console.error("Error fetching leaderboard data:", error);
      }
    }

    fetchData();
  }, []);

  // Check if leaderboard is null or empty
  const isLeaderboardEmpty = !leaderboard || leaderboard.length === 0;

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
          {isLeaderboardEmpty ? (
            <tr>
              <td colSpan="4">No data available</td>  {/* Adjusted colSpan to 4 */}
            </tr>
          ) : (
            leaderboard.map((item, index) => (
              <tr key={index}>
                <td>{item.rank}</td>
                <td>{item.name}</td>
                <td>{item.score}</td>
                <td>{item.cost}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}




function App() {
  const [loadedImageUrl, setLoadedImageUrl] = useState(null);

  useEffect(() => {
    async function fetchImage() {
      try {
        const response = await fetch(API_ENDPOINTS.getImage);
        if (!response.ok) {
          throw new Error("Failed to fetch image");
        }

        const data = await response.json();
        console.log("Image fetched on load:", data);
        setLoadedImageUrl(data.imageUrl); 
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
