import { z } from "zod";
import superjson from 'superjson';
import { type Selectable } from "kysely";
import { type Tickets, TicketStatusArrayValues, TicketPriorityArrayValues } from "../helpers/schema";

export const schema = z.object({
  status: z.enum(TicketStatusArrayValues).optional(),
  priority: z.enum(TicketPriorityArrayValues).optional(),
  sortBy: z.enum(['createdAt', 'updatedAt', 'priority', 'status']).optional(),
  sortOrder: z.enum(['asc', 'desc']).optional(),
});

export type InputType = z.infer<typeof schema>;
export type OutputType = Selectable<Tickets>[];

export const getTickets = async (params?: InputType, init?: RequestInit): Promise<OutputType> => {
  const queryParams = new URLSearchParams();
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        queryParams.append(key, String(value));
      }
    });
  }

  const result = await fetch(`/_api/tickets?${queryParams.toString()}`, {
    method: "GET",
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (!result.ok) {
    const errorObject = superjson.parse(await result.text()) as { error?: string };
    throw new Error(errorObject.error || 'An unknown error occurred');
  }
  return superjson.parse<OutputType>(await result.text());
};