import { BrowserRouter as Router, Routes, Route, useSearchParams } from 'react-router-dom';
import Recorder from './components/Recorder';
import './App.css';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Recorder />} />
      </Routes>
    </Router>
  );
}

export default App;
