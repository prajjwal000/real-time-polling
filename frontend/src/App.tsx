import { Route, Switch } from "wouter"
import { PollList } from "@/components/PollList"
import { CreatePollForm } from "@/components/CreatePollForm"
import { PollDetail } from "@/components/PollDetail"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"

export function App() {
  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-8 p-8">
      <header>
        <h1 className="text-2xl font-bold">Real-Time Polling</h1>
        <p className="text-sm text-muted-foreground">
          Vote and watch results update live
        </p>
      </header>

      <Switch>
        <Route path="/">
          <CreatePollForm onCreated={() => window.location.reload()} />
        </Route>
      </Switch>

      <Switch>
        <Route path="/">
          <Card>
            <CardHeader>
              <CardTitle>Active Polls</CardTitle>
              <CardDescription>Click a poll to vote</CardDescription>
            </CardHeader>
            <div className="px-4 pb-4">
              <PollList />
            </div>
          </Card>
        </Route>
        <Route path="/poll/:id">
          <PollDetail />
        </Route>
        <Route>
          <p className="text-muted-foreground">404 — page not found</p>
        </Route>
      </Switch>
    </div>
  )
}

export default App
