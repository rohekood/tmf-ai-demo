import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../api/client';
import type { Customer } from '../../types/customer';

// API Functions
const fetchCustomers = async (name: string): Promise<Customer[]> => {
    const params = name ? { name } : {};
    const { data } = await apiClient.get('/api/customers', { params });
    // Handle potential null/wrapping from RPC
    return Array.isArray(data) ? data : (data?.customers || []);
};

const createCustomer = async (customer: Partial<Customer>) => {
    const { data } = await apiClient.post('/api/customers', customer);
    return data;
};

export const CustomerList = () => {
    const [searchName, setSearchName] = useState('');
    const queryClient = useQueryClient();

    const { data: customers, isLoading, isError, error } = useQuery({
        queryKey: ['customers', searchName],
        queryFn: () => fetchCustomers(searchName),
    });

    const mutation = useMutation({
        mutationFn: createCustomer,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['customers'] });
        },
    });

    const handleCreateDemo = () => {
        mutation.mutate({
            name: `Demo Customer ${Math.floor(Math.random() * 1000)}`,
            status: 'Active',
            partyType: 'Individual',
        });
    };

    return (
        <div>
            <div className="input-group">
                <input
                    type="text"
                    placeholder="Search by name..."
                    value={searchName}
                    onChange={(e) => setSearchName(e.target.value)}
                />
                <button className="btn-primary" onClick={handleCreateDemo}>
                    + New Demo Customer
                </button>
            </div>

            {isLoading && <p>Loading customers...</p>}
            {isError && <p style={{ color: 'red' }}>Error: {(error as Error).message}</p>}

            <div className="grid">
                {customers?.map((customer) => (
                    <div key={customer.id} className="card">
                        <h3>{customer.name}</h3>
                        <p>Status: <span style={{ color: customer.status === 'Active' ? '#10b981' : '#f59e0b' }}>{customer.status}</span></p>
                        <small>ID: {customer.id}</small>
                    </div>
                ))}
                {!isLoading && customers?.length === 0 && (
                    <p style={{ opacity: 0.6 }}>No customers found.</p>
                )}
            </div>
        </div>
    );
};
