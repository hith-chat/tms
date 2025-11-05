import { z } from "zod";
import superjson from 'superjson';
import { type Selectable } from "kysely";
import { type Tickets, TicketStatusArrayValues, TicketPriorityArrayValues } from "../../helpers/schema";

export const schema = z.object({
  id: z.number(),
  status: z.enum(TicketStatusArrayValues).optional(),
  priority: z.enum(TicketPriorityArrayValues).optional(),
});

export type InputType = z.infer<typeof schema>;
export type OutputType = Selectable<Tickets>;

export const postTicketsUpdate = async (body: InputType, init?: RequestInit): Promise<OutputType> => {
  const validatedInput = schema.parse(body);
  const result = await fetch(`/_api/tickets/update`, {
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