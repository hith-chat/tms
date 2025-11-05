import { db } from "../../helpers/db";
import { OutputType } from "./sessions_POST.schema";
import superjson from 'superjson';
import { nanoid } from 'nanoid';

export async function handle(request: Request) {
  try {
    const sessionId = nanoid();

    const newSession = await db
      .insertInto('chatSessions')
      .values({ sessionId })
      .returningAll()
      .executeTakeFirstOrThrow();

    return new Response(superjson.stringify(newSession satisfies OutputType), {
      status: 201,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error("Error creating chat session:", error);
    const errorMessage = error instanceof Error ? error.message : "An unknown error occurred";
    return new Response(superjson.stringify({ error: errorMessage }), { status: 500 });
  }
}