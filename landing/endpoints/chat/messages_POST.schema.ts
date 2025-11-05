import { z } from "zod";
import superjson from 'superjson';
import { type Selectable } from "kysely";
import { type ChatMessages } from "../../helpers/schema";

export const schema = z.object({
  sessionId: z.string().min(1, "Session ID is required"),
  message: z.string().min(1, "Message cannot be empty"),
});

export type InputType = z.infer<typeof schema>;

export type OutputType = {
  userMessage: Selectable<ChatMessages>;
  aiMessage: Selectable<ChatMessages>;
};

export const postChatMessages = async (body: InputType, init?: RequestInit): Promise<OutputType> => {
  const validatedInput = schema.parse(body);
  const result = await fetch(`/_api/chat/messages`, {
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