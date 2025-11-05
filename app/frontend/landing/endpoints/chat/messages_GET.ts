import { db } from "../../helpers/db";
import { schema, OutputType } from "./messages_GET.schema";
import superjson from 'superjson';

export async function handle(request: Request) {
  try {
    const url = new URL(request.url);
    const sessionId = url.searchParams.get('sessionId');

    if (!sessionId) {
      return new Response(superjson.stringify({ error: "sessionId is required" }), { status: 400 });
    }

    const messages = await db
      .selectFrom('chatMessages')
      .selectAll()
      .where('sessionId', '=', sessionId)
      .orderBy('createdAt', 'asc')
      .execute();

    return new Response(superjson.stringify(messages satisfies OutputType), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error("Error fetching chat messages:", error);
    const errorMessage = error instanceof Error ? error.message : "An unknown error occurred";
    return new Response(superjson.stringify({ error: errorMessage }), { status: 500 });
  }
}