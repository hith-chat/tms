import { useState, useEffect } from 'react'
import { Users, Search, Mail, Calendar, User } from 'lucide-react'
import { apiClient, Customer, CustomersResponse } from '../lib/api'

export function CustomersPage() {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [nextCursor, setNextCursor] = useState<string | undefined>()

  const loadCustomers = async (searchQuery = '', cursor?: string) => {
    setLoading(true)
    setError(null)

    try {
      const response: CustomersResponse = await apiClient.getCustomers({
        search: searchQuery || undefined,
        cursor,
        limit: 50
      })

      if (cursor) {
        // If loading more (pagination), append to existing customers
        setCustomers(prev => [...prev, ...response.customers])
      } else {
        // New search or initial load, replace customers
        setCustomers(response.customers)
      }

      setNextCursor(response.next_cursor)
    } catch (err) {
      setError('Failed to load customers')
      console.error('Customers load error:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCustomers()
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    loadCustomers(search)
  }

  const handleLoadMore = () => {
    if (nextCursor && !loading) {
      loadCustomers(search, nextCursor)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  const getPhone = (metadata?: Record<string, string>) => {
    return metadata?.phone || 'N/A'
  }

  return (
    <div className="h-full flex flex-col bg-gradient-to-br from-background via-background to-slate-50/20 dark:to-slate-950/20">
      {/* Enhanced Header with gradient and glass effect */}
      <div className="border-b border-border/50 bg-background/80 backdrop-blur-xl supports-[backdrop-filter]:bg-background/60 shadow-sm">
        <div className="px-6 py-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-4">
              <div className="relative">
                <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg blur opacity-25"></div>
                <div className="relative p-3 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-950 dark:to-purple-950 rounded-lg border border-blue-200/50 dark:border-blue-800/50">
                  <Users className="w-6 h-6 text-blue-600 dark:text-blue-400" />
                </div>
              </div>
              <div>
                <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                  Customers
                </h1>
                <div className="flex items-center gap-3 mt-1">
                  <p className="text-sm text-muted-foreground">
                    View and manage customer information
                  </p>
                </div>
              </div>
            </div>

            {/* Search */}
            <div className="flex items-center space-x-4">
              <form onSubmit={handleSearch} className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="Search customers..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="flex h-10 w-80 rounded-lg border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-shadow"
                />
              </form>
            </div>
          </div>

          {error && (
            <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 mb-4">
              <p className="text-destructive text-sm">{error}</p>
            </div>
          )}
        </div>
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-hidden">
        <div className="h-full overflow-y-auto custom-scrollbar">
          <div className="p-6 md:p-12">
            {loading && customers.length === 0 ? (
              <div className="flex items-center justify-center py-12">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" role="status" aria-label="Loading"></div>
              </div>
            ) : (
              <div className="space-y-6">
                {/* Customers Table */}
                <div className="border rounded-lg overflow-hidden bg-card shadow-sm">
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead className="bg-muted/50">
                        <tr>
                          <th className="px-6 py-4 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                            Customer
                          </th>
                          <th className="px-6 py-4 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                            Contact Info
                          </th>
                          <th className="px-6 py-4 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                            Created
                          </th>
                          <th className="px-6 py-4 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                            Last Updated
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-background divide-y divide-border">
                        {customers.length === 0 && !loading ? (
                          <tr>
                            <td colSpan={4} className="px-6 py-12 text-center">
                              <div className="flex flex-col items-center">
                                <Users className="h-12 w-12 text-muted-foreground mb-4" />
                                <p className="text-muted-foreground">No customers found.</p>
                                {search && (
                                  <p className="text-sm text-muted-foreground mt-1">
                                    Try adjusting your search criteria.
                                  </p>
                                )}
                              </div>
                            </td>
                          </tr>
                        ) : (
                          customers.map((customer) => (
                            <tr key={customer.id} className="hover:bg-muted/50 transition-colors">
                              {/* Customer Info */}
                              <td className="px-6 py-4">
                                <div className="flex items-center space-x-3">
                                  <div className="flex-shrink-0">
                                    <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                                      <User className="h-5 w-5 text-primary" />
                                    </div>
                                  </div>
                                  <div>
                                    <div className="font-medium text-foreground">{customer.name}</div>
                                    <div className="text-sm text-muted-foreground">ID: {customer.id.slice(0, 8)}...</div>
                                  </div>
                                </div>
                              </td>

                              {/* Contact Info */}
                              <td className="px-6 py-4">
                                <div className="space-y-1">
                                  <div className="flex items-center space-x-2">
                                    <Mail className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm text-foreground">{customer.email}</span>
                                  </div>
                                  <div className="flex items-center space-x-2">
                                    <span className="h-4 w-4 text-muted-foreground text-xs">📞</span>
                                    <span className="text-sm text-muted-foreground">{getPhone(customer.metadata)}</span>
                                  </div>
                                </div>
                              </td>

                              {/* Created */}
                              <td className="px-6 py-4">
                                <div className="flex items-center space-x-2">
                                  <Calendar className="h-4 w-4 text-muted-foreground" />
                                  <span className="text-sm text-muted-foreground">{formatDate(customer.created_at)}</span>
                                </div>
                              </td>

                              {/* Last Updated */}
                              <td className="px-6 py-4">
                                <div className="flex items-center space-x-2">
                                  <Calendar className="h-4 w-4 text-muted-foreground" />
                                  <span className="text-sm text-muted-foreground">{formatDate(customer.updated_at)}</span>
                                </div>
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Load More Button */}
                {nextCursor && (
                  <div className="flex justify-center">
                    <button
                      onClick={handleLoadMore}
                      disabled={loading}
                      className="px-6 py-2 border border-border rounded-lg hover:bg-accent hover:text-accent-foreground transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {loading ? 'Loading...' : 'Load More'}
                    </button>
                  </div>
                )}

                {/* Footer Info */}
                <div className="text-center text-sm text-muted-foreground">
                  Showing {customers.length} customer{customers.length !== 1 ? 's' : ''}
                  {nextCursor && ' (more available)'}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}