function apiBase(): string {
	const override = import.meta.env.VITE_API_URL as string | undefined
	if (override !== undefined && override !== '') {
		return override.replace(/\/$/, '')
	}
	// Vite dev server → talk to local API; production build → same-origin /api (nginx proxy in Docker).
	if (import.meta.env.DEV) {
		return 'http://localhost:8080'
	}
	return ''
}

const base = apiBase()

export interface StatusChange {
  status: string
  notes?: string
  timestamp: string
}

export interface Shipment {
  id: string
  origin: string
  destination: string
  customer: string
  carrier: string
  createdAt: string
  status: string
  notes?: string
  statusHistory: StatusChange[]
}

async function readError(res: Response): Promise<string> {
  const text = await res.text()
  try {
    const j = JSON.parse(text) as { error?: string }
    if (j.error) return j.error
  } catch {
    /* not JSON */
  }
  return text || `${res.status} ${res.statusText}`
}

export async function listShipments(): Promise<Shipment[]> {
  const r = await fetch(`${base}/api/shipments`)
  if (!r.ok) throw new Error(await readError(r))
  return r.json() as Promise<Shipment[]>
}

export async function getShipment(id: string): Promise<Shipment> {
  const r = await fetch(`${base}/api/shipments/${encodeURIComponent(id)}`)
  if (!r.ok) throw new Error(await readError(r))
  return r.json() as Promise<Shipment>
}

export async function createShipment(body: {
  id: string
  origin: string
  destination: string
  customer: string
  carrier: string
}): Promise<void> {
  const r = await fetch(`${base}/api/shipments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!r.ok) throw new Error(await readError(r))
}

export async function updateStatus(id: string, status: string, notes: string): Promise<void> {
  const r = await fetch(`${base}/api/shipments/${encodeURIComponent(id)}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status, notes }),
  })
  if (!r.ok) throw new Error(await readError(r))
}
