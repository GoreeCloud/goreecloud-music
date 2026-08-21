import React from 'react'
import ReactDOM from 'react-dom/client'
import './styles.css'

function App() {
  return (
    <main className="app-shell">
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand">GoreeCloud Music</div>
        <nav>
          <a className="active" href="#home">Home</a>
          <a href="#explore">Explore</a>
          <a href="#library">Library</a>
          <a href="#search">Search</a>
        </nav>
      </aside>

      <section className="content" id="home">
        <header className="hero">
          <p className="eyebrow">Private. Personal. Yours.</p>
          <h1>Your music, on your cloud.</h1>
          <p>
            A first-party GoreeCloud music service for multi-user libraries, lossless playback,
            recommendations, radio, playlists, and offline listening.
          </p>
        </header>

        <section aria-labelledby="foundation-title">
          <h2 id="foundation-title">Foundation preview</h2>
          <div className="card-grid">
            {['Recently played', 'Made for you', 'Your library', 'Radio'].map((label) => (
              <article className="card" key={label}>
                <div className="artwork-placeholder" aria-hidden="true" />
                <strong>{label}</strong>
                <span>Coming during active development</span>
              </article>
            ))}
          </div>
        </section>
      </section>

      <footer className="player" aria-label="Now playing">
        <div>
          <strong>Nothing playing</strong>
          <span>Select music to begin</span>
        </div>
        <button type="button" disabled aria-label="Play">▶</button>
        <div className="quality">Original quality</div>
      </footer>
    </main>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
