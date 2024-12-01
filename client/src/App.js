import './App.css';


function submitPrint() {
  const submit = () => {

  }

  return (
    <button onClick={submit}>Print!</button>
  );
}


function App() {
  return (
    <div className="App">

      <div className="submit-container">
        {submitPrint()}
      </div>
    </div>
  );
}

export default App;
