import React, { useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'

function App() {
  const [tasks, setTasks] = useState([])
  const [greeting, setGreeting] = useState('Session 21 DevOps Lab')
  const [token, setToken] = useState('')
  const [title, setTitle] = useState('')
  const [filter, setFilter] = useState('all')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function refresh() {
    try {
      const response = await fetch('/api/tasks')
      if (!response.ok) throw new Error('Backend is unavailable')
      const data = await response.json()
      setTasks(data.tasks || [])
      setGreeting(data.greeting)
      setError('')
    } catch (cause) {
      setError(cause.message)
    }
  }

  useEffect(() => { refresh() }, [])

  async function write(path, method, body) {
    if (!token) { setError('Enter the API token before changing tasks.'); return false }
    setBusy(true)
    try {
      const response = await fetch(path, {
        method,
        headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined,
      })
      if (!response.ok) throw new Error(response.status === 401 ? 'Invalid API token.' : await response.text())
      await refresh()
      setError('')
      return true
    } catch (cause) {
      setError(cause.message)
      return false
    } finally {
      setBusy(false)
    }
  }

  async function create(event) {
    event.preventDefault()
    const clean = title.trim()
    if (!clean) return
    if (await write('/api/tasks', 'POST', { title: clean })) setTitle('')
  }

  const completed = tasks.filter(task => task.done).length
  const visible = useMemo(() => tasks.filter(task => filter === 'all' || (filter === 'done') === task.done), [tasks, filter])

  return <div className="app">
    <aside className="sidebar">
      <div className="brand"><span className="brand-mark">O</span><span>orbit<span className="brand-dot">.</span><small>WORKSPACE</small></span></div>
      <div className="side-label">WORKSPACE</div>
      <div className="side-link active">▦ <span>Overview</span></div>
      <div className="side-link">☷ <span>My tasks</span></div>
      <div className="side-link">◉ <span>Activity</span></div>
      <div className="side-label tools">DEVOPS TOOLS</div>
      <a className="side-link" href="/metrics" target="_blank">⌁ <span>Metrics</span><span className="out">↗</span></a>
      <a className="side-link" href="/healthz" target="_blank">♡ <span>Health check</span><span className="out">↗</span></a>
      <div className="sidebar-bottom"><span className="status-dot" /> All systems operational<br /><small>Session 21 · Final project</small></div>
    </aside>
    <main className="content">
      <header className="topbar"><span className="breadcrumb">Workspace <span>/</span> Overview</span><span className="avatar">Y</span></header>
      <div className="page">
        <div className="eyebrow">YOUR WORKSPACE <span>✦</span></div>
        <div className="heading-row"><div><h1>Good to see you<span className="accent">.</span></h1><p>{greeting} — here's what's happening with your tasks.</p></div><span className="date-pill">✦ DevOps Capstone</span></div>
        <div className="kpis">
          <article className="kpi"><span className="kpi-icon blue">▦</span><span className="kpi-label">TOTAL TASKS</span><strong>{tasks.length}</strong><small>Across your workspace</small></article>
          <article className="kpi"><span className="kpi-icon amber">◷</span><span className="kpi-label">IN PROGRESS</span><strong>{tasks.length - completed}</strong><small>Ready to be tackled</small></article>
          <article className="kpi"><span className="kpi-icon green">✓</span><span className="kpi-label">COMPLETED</span><strong>{completed}</strong><small>Nice work, keep going</small></article>
        </div>
        <section className="panel">
          <div className="panel-head"><div><h2>Task board</h2><p>Keep your work moving forward.</p></div><span className="count">{tasks.length} tasks</span></div>
          <form className="compose" onSubmit={create}><input value={title} onChange={event => setTitle(event.target.value)} maxLength="120" placeholder="Add a new task to your board..." aria-label="New task title" /><button disabled={busy} type="submit">+ Add task</button></form>
          <div className="filters">{[['all','All tasks'],['open','In progress'],['done','Completed']].map(([value,label]) => <button key={value} className={filter === value ? 'selected' : ''} onClick={() => setFilter(value)}>{label}</button>)}</div>
          <div className="task-list">{visible.length ? visible.map(task => <div className="task" key={task.id}><button className={task.done ? 'check done' : 'check'} aria-label={task.done ? 'Mark incomplete' : 'Mark complete'} disabled={busy} onClick={() => write(`/api/tasks/${task.id}`, 'PUT', { title: task.title, done: !task.done })}>{task.done ? '✓' : ''}</button><div className="task-main"><span className={task.done ? 'task-title crossed' : 'task-title'}>{task.title}</span><small>Created {new Date(task.created).toLocaleDateString()}</small></div><span className={task.done ? 'badge done' : 'badge open'}>{task.done ? 'Completed' : 'In progress'}</span><button className="delete" aria-label={`Delete ${task.title}`} disabled={busy} onClick={() => write(`/api/tasks/${task.id}`, 'DELETE')}>×</button></div>) : <div className="empty"><div>✧</div><strong>{tasks.length ? 'No tasks in this view' : 'A fresh start.'}</strong><p>{tasks.length ? 'Try another filter.' : 'Add your first task above and make things happen.'}</p></div>}</div>
        </section>
        <section className="settings"><div><strong>API access</strong><p>Enter your token to add, complete or remove tasks. It stays in this browser tab.</p></div><input value={token} onChange={event => setToken(event.target.value)} type="password" placeholder="Enter API token" autoComplete="off" aria-label="API token" /></section>
        {error && <div className="error" role="alert">{error}</div>}
        <footer>Built for Session 21 · GitHub Actions → GHCR → Kubernetes → Argo CD</footer>
      </div>
    </main>
  </div>
}

createRoot(document.getElementById('root')).render(<App />)
