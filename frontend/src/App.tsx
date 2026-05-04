import { useCallback, useEffect, useState } from 'react'
import {
  createShipment,
  getShipment,
  listShipments,
  updateStatus,
  type Shipment,
} from './api'
import './App.css'

const nextStatus = (s: string): string | null => {
  switch (s) {
    case 'CREATED':
      return 'PICKED_UP'
    case 'PICKED_UP':
      return 'IN_TRANSIT'
    case 'IN_TRANSIT':
      return 'DELIVERED'
    default:
      return null
  }
}

function App() {
  const [shipments, setShipments] = useState<Shipment[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState('')
  const [detail, setDetail] = useState<Shipment | null>(null)

  const [form, setForm] = useState({
    id: '',
    origin: '',
    destination: '',
    customer: '',
    carrier: '',
  })

  const [statusNotes, setStatusNotes] = useState('')

  const refreshList = useCallback(async () => {
    setError(null)
    const data = await listShipments()
    setShipments(data)
  }, [])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setLoading(true)
      try {
        if (!cancelled) await refreshList()
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e))
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [refreshList])

  const onCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await createShipment(form)
      setForm({ id: '', origin: '', destination: '', customer: '', carrier: '' })
      await refreshList()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const onTrack = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId.trim()) return
    setError(null)
    setLoading(true)
    try {
      const s = await getShipment(selectedId.trim())
      setDetail(s)
    } catch (err) {
      setDetail(null)
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const advanceDetail = async () => {
    if (!detail) return
    const ns = nextStatus(detail.status)
    if (!ns) return
    setError(null)
    setLoading(true)
    try {
      await updateStatus(detail.id, ns, statusNotes)
      setStatusNotes('')
      const s = await getShipment(detail.id)
      setDetail(s)
      await refreshList()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  const advanceRow = async (s: Shipment) => {
    const ns = nextStatus(s.status)
    if (!ns) return
    setError(null)
    setLoading(true)
    try {
      await updateStatus(s.id, ns, '')
      if (detail?.id === s.id) {
        setDetail(await getShipment(s.id))
      }
      await refreshList()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="app">
      <header className="header">
        <div>
          <p className="eyebrow">Hyperledger Fabric</p>
          <h1>HashedRoute</h1>
          <p className="sub">
            Shipments anchored on-chain; API serves Org1 test-network identities.
          </p>
        </div>
        <button
          type="button"
          className="btn ghost"
          disabled={loading}
          onClick={() => refreshList().catch((e) => setError(String(e)))}
        >
          Refresh ledger view
        </button>
      </header>

      {error && (
        <div className="banner error" role="alert">
          {error}
        </div>
      )}

      <div className="grid">
        <section className="card">
          <h2>New shipment</h2>
          <form className="form" onSubmit={onCreate}>
            <label>
              ID
              <input
                required
                value={form.id}
                onChange={(ev) => setForm({ ...form, id: ev.target.value })}
                placeholder="SHP-1001"
              />
            </label>
            <label>
              Origin
              <input
                required
                value={form.origin}
                onChange={(ev) => setForm({ ...form, origin: ev.target.value })}
              />
            </label>
            <label>
              Destination
              <input
                required
                value={form.destination}
                onChange={(ev) =>
                  setForm({ ...form, destination: ev.target.value })
                }
              />
            </label>
            <label>
              Customer
              <input
                required
                value={form.customer}
                onChange={(ev) => setForm({ ...form, customer: ev.target.value })}
              />
            </label>
            <label>
              Carrier
              <input
                required
                value={form.carrier}
                onChange={(ev) => setForm({ ...form, carrier: ev.target.value })}
              />
            </label>
            <button className="btn primary" type="submit" disabled={loading}>
              Create on chain
            </button>
          </form>
        </section>

        <section className="card">
          <h2>Track &amp; advance</h2>
          <form className="form inline" onSubmit={onTrack}>
            <label className="grow">
              Shipment ID
              <input
                value={selectedId}
                onChange={(ev) => setSelectedId(ev.target.value)}
                placeholder="lookup id"
              />
            </label>
            <button className="btn" type="submit" disabled={loading}>
              Load
            </button>
          </form>

          {detail && (
            <div className="detail">
              <dl>
                <div>
                  <dt>Route</dt>
                  <dd>
                    {detail.origin} → {detail.destination}
                  </dd>
                </div>
                <div>
                  <dt>Parties</dt>
                  <dd>
                    {detail.customer} · {detail.carrier}
                  </dd>
                </div>
                <div>
                  <dt>Status</dt>
                  <dd>
                    <span className={`pill st-${detail.status.toLowerCase()}`}>
                      {detail.status}
                    </span>
                  </dd>
                </div>
                <div>
                  <dt>Created</dt>
                  <dd>{detail.createdAt}</dd>
                </div>
              </dl>
              {nextStatus(detail.status) && (
                <div className="advance">
                  <label>
                    Notes (optional)
                    <input
                      value={statusNotes}
                      onChange={(ev) => setStatusNotes(ev.target.value)}
                      placeholder="dock bay, signature, etc."
                    />
                  </label>
                  <button
                    type="button"
                    className="btn primary"
                    disabled={loading}
                    onClick={() => void advanceDetail()}
                  >
                    Mark as {nextStatus(detail.status)}
                  </button>
                </div>
              )}
              {detail.statusHistory?.length > 0 && (
                <div className="history">
                  <h3>History</h3>
                  <ul>
                    {detail.statusHistory.map((h, i) => (
                      <li key={`${h.timestamp}-${i}`}>
                        <strong>{h.status}</strong>
                        <span className="muted">
                          {h.timestamp}
                          {h.notes ? ` · ${h.notes}` : ''}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </section>
      </div>

      <section className="card wide">
        <h2>All shipments</h2>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Route</th>
                <th>Status</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {shipments.length === 0 && !loading && (
                <tr>
                  <td colSpan={5} className="muted">
                    No rows yet — create a shipment or confirm the chaincode is
                    deployed.
                  </td>
                </tr>
              )}
              {shipments.map((s) => (
                <tr key={s.id}>
                  <td className="mono">{s.id}</td>
                  <td>
                    {s.origin} → {s.destination}
                  </td>
                  <td>
                    <span className={`pill st-${s.status.toLowerCase()}`}>
                      {s.status}
                    </span>
                  </td>
                  <td className="muted">{s.createdAt}</td>
                  <td className="actions">
                    {nextStatus(s.status) && (
                      <button
                        type="button"
                        className="btn small"
                        disabled={loading}
                        onClick={() => void advanceRow(s)}
                      >
                        → {nextStatus(s.status)}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  )
}

export default App
