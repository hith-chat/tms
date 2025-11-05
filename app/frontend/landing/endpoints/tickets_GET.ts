import { db } from "../helpers/db";
import { schema, OutputType } from "./tickets_GET.schema";
import superjson from 'superjson';
import { Kysely } from "kysely";
import { DB, TicketStatusArrayValues, TicketPriorityArrayValues } from "../helpers/schema";

export async function handle(request: Request) {
  try {
    const url = new URL(request.url);
    // The schema validation is for the client-side helper, 
    // but we manually handle query params on the server.
    const status = url.searchParams.get('status');
    const priority = url.searchParams.get('priority');
    const sortBy = url.searchParams.get('sortBy') || 'createdAt';
    const sortOrder = url.searchParams.get('sortOrder') || 'desc';

    let query = db.selectFrom('tickets').selectAll();

    if (status && TicketStatusArrayValues.includes(status as any)) {
      query = query.where('status', '=', status as any);
    }

    if (priority && TicketPriorityArrayValues.includes(priority as any)) {
      query = query.where('priority', '=', priority as any);
    }

    if (sortBy === 'createdAt' || sortBy === 'updatedAt' || sortBy === 'priority' || sortBy === 'status') {
        query = query.orderBy(sortBy as any, sortOrder === 'asc' ? 'asc' : 'desc');
    } else {
        // Default or invalid sort column
        query = query.orderBy('createdAt', 'desc');
    }

    const tickets = await query.execute();

    return new Response(superjson.stringify(tickets satisfies OutputType), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error("Error fetching tickets:", error);
    const errorMessage = error instanceof Error ? error.message : "An unknown error occurred";
    return new Response(superjson.stringify({ error: errorMessage }), { status: 500 });
  }
}