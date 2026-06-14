import { useState, useMemo, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
    createColumnHelper,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
    flexRender,
    type SortingState,
} from '@tanstack/react-table';
import { useReactTable } from '../../hooks/useReactTable';
import { Plus, Search, User, Building2, ChevronUp, ChevronDown, Eye, Edit, Trash2, Loader2 } from 'lucide-react';
import { EmptyState } from '../../design-system/components/common/EmptyState';
import { useParties, useDeleteParty, usePurgeParty } from './api';
import { type PartyUnion, getPartyDisplayName, isIndividual } from './types';
import { useNotification } from '../../design-system/components/common/Toast';
import { IconButton } from '../../design-system/components/common/IconButton';
import { IconButtonArea } from '../../design-system/components/common/IconButtonArea';
import { useQueryClient } from '@tanstack/react-query';
import './PartyListPage.css';

const columnHelper = createColumnHelper<PartyUnion>();

export default function PartyListPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);
    const [showDeleted, setShowDeleted] = useState(false);
    const queryClient = useQueryClient();

    const searchQuery = searchParams.get('q') || '';

    const queryParams = {
        ...(searchQuery ? { search: searchQuery } : {}),
        ...(!showDeleted ? { status: 'Active' } : {}),
    };

    const { data: parties = [], isLoading, error } = useParties(
        Object.keys(queryParams).length > 0 ? queryParams : undefined
    );

    const deleteMutation = useDeleteParty();
    const purgeMutation = usePurgeParty();

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        if (searchTerm) {
            setSearchParams({ q: searchTerm });
        } else {
            setSearchParams({});
        }
    };

    const { showToast } = useNotification();
    const [deletingId, setDeletingId] = useState<string | null>(null);

    const checkDeletionStatus = useCallback(async (id: string, attempts = 0) => {
        if (attempts > 10) {
            setDeletingId(null);
            showToast('Deletion operation timed out. Please check status manually.', 'info');
            return;
        }

        try {
            // Fetch fresh party data via apiClient (has correct BFF URL)
            const { apiClient } = await import('../../api/client');
            const response = await apiClient.get(`/api/parties/${id}`);
            const party = response.data;

            if (party.status === 'Deleted') {
                setDeletingId(null);
                showToast('Party deleted successfully', 'success');
                queryClient.invalidateQueries({ queryKey: ['parties'] });
            } else if (party.status === 'Active') {
                setDeletingId(null);
                showToast('Deletion failed: Party has active linked customers.', 'error');
                queryClient.invalidateQueries({ queryKey: ['parties'] });
            } else if (party.status === 'DeletionPending') {
                setTimeout(() => checkDeletionStatus(id, attempts + 1), 1000);
            } else {
                setDeletingId(null);
                showToast(`Deletion ended with status: ${party.status}`, 'info');
                queryClient.invalidateQueries({ queryKey: ['parties'] });
            }

        } catch (err: unknown) {
            // 404 means party was deleted
            if (err && typeof err === 'object' && 'response' in err) {
                const axiosErr = err as { response?: { status?: number } };
                if (axiosErr.response?.status === 404) {
                    setDeletingId(null);
                    showToast('Party deleted successfully', 'success');
                    queryClient.invalidateQueries({ queryKey: ['parties'] });
                    return;
                }
            }
            console.error("Error checking status", err);
            setDeletingId(null);
            showToast('Error checking deletion status', 'error');
        }
    }, [showToast, queryClient]);

    const handleDelete = useCallback((id: string, name: string) => {
        if (confirm(`Are you sure you want to delete "${name}"?`)) {
            deleteMutation.mutate(id, {
                onSuccess: () => {
                    showToast('Deletion initiated...', 'info');
                    setDeletingId(id);
                    checkDeletionStatus(id);
                },
                onError: (err: unknown) => {
                    const axiosErr = err as { response?: { status?: number; data?: string } };
                    if (axiosErr.response?.status === 409) {
                        showToast(axiosErr.response.data || 'Cannot delete: party has linked customers.', 'error');
                    } else {
                        showToast(`Failed to initiate deletion: ${(err as Error).message}`, 'error');
                    }
                }
            });
        }
    }, [deleteMutation, showToast, checkDeletionStatus]);

    const handlePurge = useCallback((id: string, name: string) => {
        if (confirm(`Permanently delete "${name}"? This cannot be undone.`)) {
            purgeMutation.mutate(id, {
                onSuccess: () => {
                    showToast('Party permanently deleted', 'success');
                    queryClient.invalidateQueries({ queryKey: ['parties'] });
                },
                onError: (err) => {
                    showToast(`Failed to permanently delete: ${err.message}`, 'error');
                }
            });
        }
    }, [purgeMutation, showToast, queryClient]);

    const columns = useMemo(
        () => [
            columnHelper.accessor((row) => row['@type'], {
                id: 'type',
                header: 'Type',
                cell: (info) => {
                    const value = info.getValue();
                    const badgeClass = value ? value.toLowerCase() : 'unknown';
                    return (
                        <button
                            type="button"
                            className={`party-type-badge ${badgeClass}`}
                            aria-label={`Filter by ${value}`}
                        >
                            {value === 'Individual' ? <User size={14} /> : <Building2 size={14} />}
                            {value}
                        </button>
                    );
                },
                size: 140,
            }),
            columnHelper.accessor((row) => getPartyDisplayName(row), {
                id: 'name',
                header: 'Name',
                cell: (info) => <span className="party-name">{info.getValue()}</span>,
            }),
            columnHelper.accessor('status', {
                header: 'Status',
                cell: (info) => (
                    <span className={`status-badge ${info.getValue()}`}>
                        {info.getValue()}
                    </span>
                ),
                size: 100,
            }),
            columnHelper.accessor((row) => (isIndividual(row) ? row.givenName : row.tradingName), {
                id: 'identifier',
                header: 'Identifier',
                cell: (info) => {
                    const party = info.row.original;
                    const identifications = party.identifications || [];
                    if (identifications.length > 0) {
                        return (
                            <span className="identifier-text">
                                {identifications[0].identificationType}: {identifications[0].identificationId}
                            </span>
                        );
                    }
                    return <span className="text-muted">No ID</span>;
                },
                size: 180,
            }),
            columnHelper.display({
                id: 'actions',
                header: '',
                cell: (info) => {
                    const party = info.row.original;
                    const isDeleted = party.status === 'Deleted';
                    return (
                        <IconButtonArea alignment="end">
                            <IconButton
                                to={`/parties/${party.id}`}
                                icon={<Eye size={16} />}
                                title="View"
                            />
                            {!isDeleted && (
                                <IconButton
                                    to={`/parties/${party.id}/edit`}
                                    icon={<Edit size={16} />}
                                    title="Edit"
                                />
                            )}
                            {isDeleted ? (
                                <IconButton
                                    variant="danger"
                                    title="Permanently Delete"
                                    onClick={() => handlePurge(party.id, getPartyDisplayName(party))}
                                    disabled={purgeMutation.isPending}
                                    icon={<Trash2 size={16} />}
                                />
                            ) : (
                                <IconButton
                                    variant="danger"
                                    title="Delete"
                                    onClick={() => handleDelete(party.id, getPartyDisplayName(party))}
                                    disabled={deleteMutation.isPending || deletingId === party.id}
                                    icon={deletingId === party.id ? <Loader2 size={16} className="spin" /> : <Trash2 size={16} />}
                                />
                            )}
                        </IconButtonArea>
                    );
                },
                size: 140,
            }),
        ],
        [deleteMutation.isPending, purgeMutation.isPending, handleDelete, handlePurge, deletingId]
    );

    const table = useReactTable({
        data: parties,
        columns,
        state: { sorting },
        onSortingChange: setSorting,
        getCoreRowModel: getCoreRowModel(),
        getSortedRowModel: getSortedRowModel(),
        getFilteredRowModel: getFilteredRowModel(),
    });

    return (
        <div className="party-list-page">
            <div className="page-header">
                <div className="page-header-content">
                    <h1 className="page-title">Parties</h1>
                    <p className="page-description">Manage Individuals and Organizations</p>
                </div>
                <Link to="/parties/new" className="btn btn-primary">
                    <Plus size={18} />
                    <span>Add Party</span>
                </Link>
            </div>

            <div className="search-bar card" style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <Search size={20} className="search-icon" />
                <form style={{ display: 'contents' }} onSubmit={handleSearch}>
                    <input
                        type="text"
                        placeholder="Search by name, ID, or identification..."
                        className="search-input"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                    <button type="submit" className="btn btn-secondary">
                        Search
                    </button>
                </form>
                <label style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', whiteSpace: 'nowrap', fontSize: '0.875rem', cursor: 'pointer' }}>
                    <input
                        type="checkbox"
                        checked={showDeleted}
                        onChange={(e) => setShowDeleted(e.target.checked)}
                    />
                    Show deleted
                </label>
            </div>

            {error ? (
                <div className="card error-card" role="alert">
                    <p>Failed to load parties: {error.message}</p>
                </div>
            ) : isLoading ? (
                <div className="card loading-card" role="status">
                    <Loader2 className="spin" size={24} />
                    <p>Loading parties...</p>
                </div>
            ) : parties.length === 0 ? (
                <EmptyState
                    icon={<User size={48} />}
                    title="No parties found."
                    description="Create an individual or organization to get started."
                    action={
                        <Link to="/parties/new" className="btn btn-primary">
                            Create your first party
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
                        <span className="table-count">{parties.length} parties</span>
                    </div>
                </div>
            )}
        </div>
    );
}
