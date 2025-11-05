import { db } from "../helpers/db";
import { schema, OutputType } from "./tickets_POST.schema";
import superjson from 'superjson';

export async function handle(request: Request) {
  try {
    const json = superjson.parse(await request.text());
    const input = schema.parse(json);

    const newTicket = await db
      .insertInto('tickets')
      .values({
        title: input.title,
        description: input.description,
        priority: input.priority,
        status: 'open', // New tickets always start as 'open'
      })
      .returningAll()
      .executeTakeFirstOrThrow();

    return new Response(superjson.stringify(newTicket satisfies OutputType), {
      status: 201,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error("Error creating ticket:", error);
    const errorMessage = error instanceof Error ? error.message : "An unknown error occurred";
    return new Response(superjson.stringify({ error: errorMessage }), { status: 400 });
  }
}