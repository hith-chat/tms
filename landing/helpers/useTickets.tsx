import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getTickets, type InputType as GetTicketsInput } from '../endpoints/tickets_GET.schema';
import { postTickets, type InputType as CreateTicketInput } from '../endpoints/tickets_POST.schema';
import { postTicketsUpdate, type InputType as UpdateTicketInput } from '../endpoints/tickets/update_POST.schema';

export const useTicketsQueryKey = (filters: GetTicketsInput = {}) => ['tickets', filters];

export const useTickets = (filters: GetTicketsInput = {}) => {
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: useTicketsQueryKey(filters),
    queryFn: () => getTickets(filters),
  });

  const createTicketMutation = useMutation({
    mutationFn: (newTicket: CreateTicketInput) => postTickets(newTicket),
    onSuccess: () => {
      // Invalidate all ticket queries to refetch the list with the new ticket
      queryClient.invalidateQueries({ queryKey: ['tickets'] });
    },
  });

  const updateTicketMutation = useMutation({
    mutationFn: (updatedTicket: UpdateTicketInput) => postTicketsUpdate(updatedTicket),
    onSuccess: () => {
      // Invalidate all ticket queries as status/priority changes might affect filtering
      queryClient.invalidateQueries({ queryKey: ['tickets'] });
    },
  });

  return {
    ...query,
    createTicket: createTicketMutation.mutate,
    createTicketAsync: createTicketMutation.mutateAsync,
    isCreatingTicket: createTicketMutation.isPending,
    updateTicket: updateTicketMutation.mutate,
    updateTicketAsync: updateTicketMutation.mutateAsync,
    isUpdatingTicket: updateTicketMutation.isPending,
  };
};