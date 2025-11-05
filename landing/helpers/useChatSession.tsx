import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { postChatSessions } from '../endpoints/chat/sessions_POST.schema';
import { getChatMessages } from '../endpoints/chat/messages_GET.schema';
import { postChatMessages, type InputType as PostMessageInput } from '../endpoints/chat/messages_POST.schema';
import { type Selectable } from 'kysely';
import { type ChatMessages } from './schema';

export const useChatMessagesQueryKey = (sessionId: string | null | undefined) => ['chatMessages', sessionId];

export const useChatSession = (sessionId: string | null | undefined) => {
  const queryClient = useQueryClient();

  const messagesQuery = useQuery({
    queryKey: useChatMessagesQueryKey(sessionId),
    queryFn: () => getChatMessages({ sessionId: sessionId! }),
    enabled: !!sessionId, // Only fetch if sessionId is available
  });

  const createSessionMutation = useMutation({
    mutationFn: postChatSessions,
  });

  const sendMessageMutation = useMutation({
    mutationFn: (newMessage: PostMessageInput) => postChatMessages(newMessage),
    onSuccess: (data) => {
      // Optimistically update the messages list
      queryClient.setQueryData<Selectable<ChatMessages>[]>(
        useChatMessagesQueryKey(data.userMessage.sessionId),
        (oldData) => {
          if (!oldData) return [data.userMessage, data.aiMessage];
          return [...oldData, data.userMessage, data.aiMessage];
        }
      );
    },
  });

  return {
    messages: messagesQuery.data,
    isFetchingMessages: messagesQuery.isFetching,
    messagesError: messagesQuery.error,
    createSession: createSessionMutation.mutate,
    createSessionAsync: createSessionMutation.mutateAsync,
    isCreatingSession: createSessionMutation.isPending,
    sendMessage: sendMessageMutation.mutate,
    sendMessageAsync: sendMessageMutation.mutateAsync,
    isSendingMessage: sendMessageMutation.isPending,
  };
};