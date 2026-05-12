import { useEffect, useState, useCallback } from "react"
import { useParams, Link } from "wouter"
import { getPoll, vote, getResults, hasVoted, type Poll, type VoteCount } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card"
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart"
import { Bar, BarChart, XAxis, YAxis } from "recharts"

export function PollDetail() {
  const { id } = useParams()
  const [poll, setPoll] = useState<Poll | null>(null)
  const [results, setResults] = useState<VoteCount[]>([])
  const [voted, setVoted] = useState(false)
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(true)

  const pollId = Number(id)

  const fetchResults = useCallback(async () => {
    try {
      const data = await getResults(pollId)
      setResults(data)
    } catch (err) {
      console.error(err)
    }
  }, [pollId])

  useEffect(() => {
    getPoll(pollId)
      .then(setPoll)
      .catch(() => setError("Poll not found"))
      .finally(() => setLoading(false))
    hasVoted(pollId).then(setVoted).catch(console.error)
  }, [pollId])

  useEffect(() => {
    if (!poll) return
    fetchResults()
    const interval = setInterval(fetchResults, 2000)
    return () => clearInterval(interval)
  }, [poll, fetchResults])

  const handleVote = async (optionId: number) => {
    try {
      await vote(pollId, optionId)
      setVoted(true)
      fetchResults()
    } catch (err: unknown) {
      if (err instanceof Error && err.message === "already-voted") {
        setVoted(true)
      } else {
        setError("Failed to submit vote")
      }
    }
  }

  if (loading) return <p className="text-muted-foreground">Loading...</p>
  if (error) return <p className="text-destructive">{error}</p>
  if (!poll) return null

  const optionCounts: Record<string, number> = {}
  const chartConfig: ChartConfig = {}

  for (const opt of poll.options ?? []) {
    const key = opt.id.toString()
    const result = results.find((r) => r.option_id === opt.id)
    optionCounts[key] = result?.count ?? 0
    chartConfig[key] = { label: opt.text }
  }

  const chartData = poll.options?.map((opt) => ({
    option: opt.text,
    votes: optionCounts[opt.id.toString()] ?? 0,
    fill: `hsl(${opt.id * 60}, 70%, 50%)`,
  })) ?? []

  return (
    <div className="flex flex-col gap-6">
      <Link href="/" className="text-sm text-muted-foreground hover:underline">
        ← Back to polls
      </Link>

      <Card>
        <CardHeader>
          <CardTitle>{poll.question}</CardTitle>
        </CardHeader>
        <CardContent>
          {voted ? (
            <p className="text-sm text-muted-foreground">You voted on this poll.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {poll.options?.map((opt) => (
                <Button key={opt.id} variant="outline" onClick={() => handleVote(opt.id)}>
                  {opt.text}
                </Button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Live Results</CardTitle>
        </CardHeader>
        <CardContent>
          <ChartContainer config={chartConfig} className="aspect-auto h-[400px]">
            <BarChart data={chartData} accessibilityLayer>
              <XAxis dataKey="option" tickLine={false} tickMargin={10} />
              <YAxis allowDecimals={false} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Bar dataKey="votes" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  )
}
