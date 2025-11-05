import { z } from "zod";
import superjson from 'superjson';
import { type Selectable } from "kysely";
import { type ChatSessions } from "../../helpers/schema";

export const schema = z.object({});

export type InputType = z.infer<typeof schema>;
export type OutputType = Selectable<ChatSessions>;

export const postChatSessions = async (init?: RequestInit): Promise<OutputType> => {
  const result = await fetch(`/_api/chat/sessions`, {
    method: "POST",
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