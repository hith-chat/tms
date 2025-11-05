import { z } from "zod";
import superjson from 'superjson';
import { type Selectable } from "kysely";
import { type ChatMessages } from "../../helpers/schema";

export const schema = z.object({
  sessionId: z.string(),
});

export type InputType = z.infer<typeof schema>;
export type OutputType = Selectable<ChatMessages>[];

export const getChatMessages = async (params: InputType, init?: RequestInit): Promise<OutputType> => {
  const validatedParams = schema.parse(params);
  const queryParams = new URLSearchParams({ sessionId: validatedParams.sessionId });

  const result = await fetch(`/_api/chat/messages?${queryParams.toString()}`, {
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