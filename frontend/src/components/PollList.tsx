import { useEffect, useState } from "react"
import { Link } from "wouter"
import { listPolls, type Poll } from "@/lib/api"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"

export function PollList() {
  const [polls, setPolls] = useState<Poll[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listPolls()
      .then(setPolls)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <p className="text-muted-foreground">Loading polls...</p>

  if (polls.length === 0) return <p className="text-muted-foreground">No polls yet. Create one above!</p>

  return (
    <div className="grid gap-3">
      {polls.map((poll) => (
        <Link key={poll.id} href={`/poll/${poll.id}`}>
          <Card className="cursor-pointer transition-colors hover:bg-muted/50">
            <CardHeader>
              <CardTitle>{poll.question}</CardTitle>
              <CardDescription>
                Created {new Date(poll.created_at).toLocaleDateString()}
              </CardDescription>
            </CardHeader>
          </Card>
        </Link>
      ))}
    </div>
  )
}
