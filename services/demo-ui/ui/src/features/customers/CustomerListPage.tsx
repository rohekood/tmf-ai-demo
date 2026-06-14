import { useState, useMemo, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
    createColumnHelper,
    getCoreRowModel,
    getSortedRowModel,
    flexRender,
    type SortingState,
} from '@tanstack/react-table';
import { useReactTable } from '../../hooks/useReactTable';
import { Plus, Search, ChevronUp, ChevronDown, Eye, Edit, Trash2, Loader2, ExternalLink, UserPlus } from 'lucide-react';
import { EmptyState } from '../../design-system/components/common/EmptyState';
import { useCustomers, useDeleteCustomer } from './api';
import type { Customer } from './types';
import '../parties/PartyListPage.css'; // Reuse party list styles
import './CustomerListPage.css';
import { IconButton } from '../../design-system/components/common/IconButton';
import { IconButtonArea } from '../../design-system/components/common/IconButtonArea';

const columnHelper = createColumnHelper<Customer>();

export default function CustomerListPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);

    const searchQuery = searchParams.get('q') || '';

    const { data: customers = [], isLoading, error } = useCustomers(
        searchQuery ? { search: searchQuery } : undefined
    );

    const deleteMutation = useDeleteCustomer();

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        if (searchTerm) {
            setSearchParams({ q: searchTerm });
        } else {
            setSearchParams({});
        }
    };

    const handleDelete = useCallback((id: string, name: string) => {
        if (confirm(`Are you sure you want to delete customer "${name}"?`)) {
            deleteMutation.mutate(id);
        }
    }, [deleteMutation]);

    const columns = useMemo(
        () => [
            columnHelper.accessor('name', {
                header: 'Name',
                cell: (info) => <span className="customer-name">{info.getValue()}</span>,
            }),
            columnHelper.accessor('status', {
                header: 'Status',
                cell: (info) => (
                    <span className={`status-badge ${info.getValue().toLowerCase()}`}>
                        {info.getValue()}
                    </span>
                ),
                size: 120,
            }),
            columnHelper.accessor('partyId', {
                header: 'Party Reference',
                cell: (info) => (
                    <Link to={`/parties/${info.getValue()}`} className="party-link">
                        {info.row.original.partyName || info.getValue().slice(0, 8) + '...'}
                        <ExternalLink size={12} />
                    </Link>
                ),
                size: 180,
            }),
            columnHelper.accessor((row) => row.accounts?.length || 0, {
                id: 'accounts',
                header: 'Accounts',
                cell: (info) => (
                    <span className="count-text">{info.getValue()} accounts</span>
                ),
                size: 100,
            }),
            columnHelper.display({
                id: 'actions',
                header: '',
                cell: (info) => (
                    <IconButtonArea alignment="end">
                        <IconButton
                            to={`/customers/${info.row.original.id}`}
                            icon={<Eye size={16} />}
                            title="View"
                        />
                        <IconButton
                            to={`/customers/${info.row.original.id}/edit`}
                            icon={<Edit size={16} />}
                            title="Edit"
                        />
                        <IconButton
                            variant="danger"
                            title="Delete"
                            onClick={() => handleDelete(info.row.original.id, info.row.original.name)}
                            disabled={deleteMutation.isPending}
                            icon={<Trash2 size={16} />}
                        />
                    </IconButtonArea>
                ),
                size: 120,
            }),
        ],
        [deleteMutation.isPending, handleDelete]
    );

    const table = useReactTable({
        data: customers,
        columns,
        state: { sorting },
        onSortingChange: setSorting,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
    });

    return (
        <div className="customer-list-page">
            <div className="page-header">
                <div className="page-header-content">
                    <h1 className="page-title">Customers</h1>
                    <p className="page-description">Manage customer relationships</p>
                </div>
                <Link to="/customers/new" className="btn btn-primary">
                    <Plus size={18} />
                    <span>Onboard Customer</span>
                </Link>
            </div>

            <form className="search-bar card" onSubmit={handleSearch}>
                <Search size={20} className="search-icon" />
                <input
                    type="text"
                    placeholder="Search by name, status, or party reference..."
                    className="search-input"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
                <button type="submit" className="btn btn-secondary">
                    Search
                </button>
            </form>

            {error ? (
                <div className="card error-card" role="alert">
                    <p>Failed to load customers: {error.message}</p>
                </div>
            ) : isLoading ? (
                <div className="card loading-card" role="status">
                    <Loader2 className="spin" size={24} />
                    <p>Loading customers...</p>
                </div>
            ) : customers.length === 0 ? (
                <EmptyState
                    icon={<UserPlus size={48} />}
                    title="No customers found."
                    description="Onboard a customer to manage their relationships."
                    action={
                        <Link to="/customers/new" className="btn btn-primary">
                            Onboard your first customer
                        </Link>
                    }
                />
            ) : (
                <div className="card table-card">
                    <table className="data-table">
                        <thead>
                            {table.getHeaderGroups().map((headerGroup) => (
                                <tr key={headerGroup.id}>
                                    {headerGroup.headers.map((header) => (
                                        <th
                                            key={header.id}
                                            onClick={header.column.getToggleSortingHandler()}
                                            className={header.column.getCanSort() ? 'sortable' : ''}
                                            style={{ width: header.getSize() }}
                                        >
                                            <div className="th-content">
                                                {flexRender(header.column.columnDef.header, header.getContext())}
                                                {header.column.getIsSorted() === 'asc' && <ChevronUp size={14} />}
                                                {header.column.getIsSorted() === 'desc' && <ChevronDown size={14} />}
                                            </div>
                                        </th>
                                    ))}
                                </tr>
                            ))}
                        </thead>
                        <tbody>
                            {table.getRowModel().rows.map((row) => (
                                <tr key={row.id}>
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id}>
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                    <div className="table-footer">
                        <span className="table-count">{customers.length} customers</span>
                    </div>
                </div>
            )}
        </div>
    );
}
