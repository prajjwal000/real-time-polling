const BASE = "/api"

export interface Poll {
  id: number
  question: string
  created_at: string
  options?: Option[]
}

export interface Option {
  id: number
  poll_id: number
  text: string
}

export interface VoteCount {
  option_id: number
  count: number
}

export async function listPolls(): Promise<Poll[]> {
  const res = await fetch(`${BASE}/polls`)
  if (!res.ok) throw new Error("Failed to fetch polls")
  return res.json()
}

export async function getPoll(id: number): Promise<Poll> {
  const res = await fetch(`${BASE}/polls/${id}`)
  if (!res.ok) throw new Error("Poll not found")
  return res.json()
}

export async function createPoll(question: string, options: string[]): Promise<Poll> {
  const res = await fetch(`${BASE}/polls`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ question, options }),
  })
  if (!res.ok) throw new Error("Failed to create poll")
  return res.json()
}

export async function vote(pollId: number, optionId: number): Promise<void> {
  const res = await fetch(`${BASE}/polls/${pollId}/vote`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ option_id: optionId }),
  })
  if (res.status === 409) throw new Error("already-voted")
  if (!res.ok) throw new Error("Failed to vote")
}

export async function hasVoted(pollId: number): Promise<boolean> {
  const res = await fetch(`${BASE}/polls/${pollId}/has-voted`)
  if (!res.ok) return false
  const data = await res.json()
  return data.voted
}

export async function getResults(pollId: number): Promise<VoteCount[]> {
  const res = await fetch(`${BASE}/polls/${pollId}/results`)
  if (!res.ok) throw new Error("Failed to fetch results")
  return res.json()
}
