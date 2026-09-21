import React, { useState } from 'react';
import './App.css';

function App() {
  const [longUrl, setLongUrl] = useState('');
  const [shortUrl, setShortUrl] = useState('');
  const [shortCode, setShortCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [stats, setStats] = useState(null);
  const [showStats, setShowStats] = useState(false);

  const shortenUrl = async () => {
    if (!longUrl.trim()) {
      setError('Please enter a URL');
      return;
    }

    setLoading(true);
    setError('');
    setShortUrl('');
    setStats(null);

    try {
      const response = await fetch('http://localhost:8080/api/v1/shorten', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ long_url: longUrl }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to shorten URL');
      }

      setShortUrl(data.short_url);
      setShortCode(data.short_code);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const getStats = async () => {
    if (!shortCode) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/stats/${shortCode}`);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to fetch stats');
      }

      setStats(data);
      setShowStats(true);
    } catch (err) {
      setError(err.message);
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(shortUrl);
    alert('Copied to clipboard!');
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>URL Shortener</h1>
        <p className="subtitle">Paste a long URL and get a short one</p>

        <div className="input-group">
          <input
            type="url"
            placeholder="https://example.com/your-very-long-url..."
            value={longUrl}
            onChange={(e) => setLongUrl(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && shortenUrl()}
            className="url-input"
          />
          <button
            onClick={shortenUrl}
            disabled={loading}
            className="shorten-btn"
          >
            {loading ? 'Shortening...' : 'Shorten'}
          </button>
        </div>

        {error && <div className="error">{error}</div>}

        {shortUrl && (
          <div className="result">
            <p className="result-label">Your shortened URL:</p>
            <div className="result-group">
              <a href={shortUrl} target="_blank" rel="noopener noreferrer" className="short-url">
                {shortUrl}
              </a>
              <button onClick={copyToClipboard} className="copy-btn">Copy</button>
              <button onClick={getStats} className="stats-btn">Stats</button>
            </div>
          </div>
        )}

        {showStats && stats && (
          <div className="stats-panel">
            <h3>Statistics for {stats.stats.short_code}</h3>
            <p className="stats-info"><strong>Original URL:</strong> {stats.stats.long_url}</p>
            <p className="stats-info"><strong>Total Clicks:</strong> {stats.stats.clicks}</p>
            <p className="stats-info"><strong>Created:</strong> {new Date(stats.stats.created_at).toLocaleString()}</p>

            {stats.analytics && stats.analytics.length > 0 && (
              <div className="analytics">
                <h4>Recent Clicks</h4>
                <table>
                  <thead>
                    <tr>
                      <th>Time</th>
                      <th>IP</th>
                      <th>Referrer</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stats.analytics.map((entry, index) => (
                      <tr key={index}>
                        <td>{new Date(entry.clicked_at).toLocaleString()}</td>
                        <td>{entry.ip_address}</td>
                        <td>{entry.referrer || 'Direct'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            <button onClick={() => setShowStats(false)} className="close-btn">Close</button>
          </div>
        )}
      </header>
    </div>
  );
}

export default App;
