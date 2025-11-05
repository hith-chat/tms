import { z } from "zod";
import superjson from 'superjson';
import { type Selectable } from "kysely";
import { type Tickets, TicketPriorityArrayValues } from "../helpers/schema";

export const schema = z.object({
  title: z.string().min(1, "Title is required").max(255),
  description: z.string().min(1, "Description is required"),
  priority: z.enum(TicketPriorityArrayValues).default('medium'),
});

export type InputType = z.infer<typeof schema>;
export type OutputType = Selectable<Tickets>;

export const postTickets = async (body: InputType, init?: RequestInit): Promise<OutputType> => {
  const validatedInput = schema.parse(body);
  const result = await fetch(`/_api/tickets`, {
    method: "POST",
    body: superjson.stringify(validatedInput),
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