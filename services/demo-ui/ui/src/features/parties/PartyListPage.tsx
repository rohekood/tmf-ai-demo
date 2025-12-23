import { useState, useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
    createColumnHelper,
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
    getFilteredRowModel,
    useReactTable,
    type SortingState,
} from '@tanstack/react-table';
import { Plus, Search, User, Building2, ChevronUp, ChevronDown, Eye, Edit, Trash2, Loader2 } from 'lucide-react';
import { useParties, useDeleteParty } from './api';
import { type PartyUnion, getPartyDisplayName, isIndividual } from './types';
import './PartyListPage.css';

const columnHelper = createColumnHelper<PartyUnion>();

export default function PartyListPage() {
    const [searchParams, setSearchParams] = useSearchParams();
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);

    const searchQuery = searchParams.get('q') || '';

    const { data: parties = [], isLoading, error } = useParties(
        searchQuery ? { givenName: searchQuery, tradingName: searchQuery } : undefined
    );

    const deleteMutation = useDeleteParty();

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        if (searchTerm) {
            setSearchParams({ q: searchTerm });
        } else {
            setSearchParams({});
        }
    };

    const handleDelete = (id: string, name: string) => {
        if (confirm(`Are you sure you want to delete "${name}"?`)) {
            deleteMutation.mutate(id);
        }
    };

    const columns = useMemo(
        () => [
            columnHelper.accessor((row) => row['@type'], {
                id: 'type',
                header: 'Type',
                cell: (info) => (
                    <span className={`party-type-badge ${info.getValue().toLowerCase()}`}>
                        {info.getValue() === 'Individual' ? <User size={14} /> : <Building2 size={14} />}
                        {info.getValue()}
                    </span>
                ),
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
                cell: (info) => (
                    <div className="action-buttons">
                        <Link to={`/parties/${info.row.original.id}`} className="action-btn" title="View">
                            <Eye size={16} />
                        </Link>
                        <Link to={`/parties/${info.row.original.id}/edit`} className="action-btn" title="Edit">
                            <Edit size={16} />
                        </Link>
                        <button
                            className="action-btn action-btn--danger"
                            title="Delete"
                            onClick={() => handleDelete(info.row.original.id, getPartyDisplayName(info.row.original))}
                            disabled={deleteMutation.isPending}
                        >
                            <Trash2 size={16} />
                        </button>
                    </div>
                ),
                size: 120,
            }),
        ],
        [deleteMutation.isPending]
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
                    <h2>Parties</h2>
                    <p className="page-description">Manage Individuals and Organizations</p>
                </div>
                <Link to="/parties/new" className="btn btn-primary">
                    <Plus size={18} />
                    <span>Add Party</span>
                </Link>
            </div>

            <form className="search-bar card" onSubmit={handleSearch}>
                <Search size={20} className="search-icon" />
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
                <div className="card empty-card" role="status">
                    <p>No parties found.</p>
                    <Link to="/parties/new" className="btn btn-primary">
                        Create your first party
                    </Link>
                </div>
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
