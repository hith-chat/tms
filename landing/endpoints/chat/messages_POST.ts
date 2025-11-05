import { db } from "../../helpers/db";
import { schema, OutputType } from "./messages_POST.schema";
import superjson from 'superjson';
import { Transaction } from "kysely";
import { DB } from "../../helpers/schema";

// A simple AI response simulator
const getAIResponse = (userMessage: string): string => {
  const lowerCaseMessage = userMessage.toLowerCase();
  if (lowerCaseMessage.includes("password") || lowerCaseMessage.includes("reset")) {
    return "If you need to reset your password, please visit the 'Forgot Password' link on the login page. We'll send you an email with instructions.";
  }
  if (lowerCaseMessage.includes("billing") || lowerCaseMessage.includes("invoice")) {
    return "You can view your billing history and manage your subscription in the 'Billing' section of your account settings.";
  }
  if (lowerCaseMessage.includes("contact") || lowerCaseMessage.includes("support")) {
    return "You can create a support ticket on our support page for detailed assistance. Our team will get back to you as soon as possible.";
  }
  if (lowerCaseMessage.includes("hello") || lowerCaseMessage.includes("hi")) {
    return "Hello! How can I assist you today?";
  }
  return "I'm sorry, I'm not sure how to help with that. Could you please rephrase your question? For more complex issues, I recommend creating a support ticket.";
};

export async function handle(request: Request) {
  try {
    const json = superjson.parse(await request.text());
    const input = schema.parse(json);

    const response = await db.transaction().execute(async (trx) => {
      const userMessage = await trx
        .insertInto('chatMessages')
        .values({
          sessionId: input.sessionId,
          message: input.message,
          role: 'user',
        })
        .returningAll()
        .executeTakeFirstOrThrow();

      const aiResponseMessage = getAIResponse(input.message);

      const aiMessage = await trx
        .insertInto('chatMessages')
        .values({
          sessionId: input.sessionId,
          message: aiResponseMessage,
          role: 'assistant',
        })
        .returningAll()
        .executeTakeFirstOrThrow();
      
      return { userMessage, aiMessage };
    });

    return new Response(superjson.stringify(response satisfies OutputType), {
      status: 201,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (error) {
    console.error("Error posting chat message:", error);
    const errorMessage = error instanceof Error ? error.message : "An unknown error occurred";
    return new Response(superjson.stringify({ error: errorMessage }), { status: 400 });
  }
}