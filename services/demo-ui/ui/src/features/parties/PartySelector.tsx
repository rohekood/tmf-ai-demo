import { useState, useMemo } from 'react';
import { Search, Loader2, Check, User, Building2, ChevronUp, ChevronDown } from 'lucide-react';
import {
    createColumnHelper,
    getCoreRowModel,
    getFilteredRowModel,
    getSortedRowModel,
    flexRender,
    useReactTable,
    type SortingState,
} from '@tanstack/react-table';
import { useParties } from './api';
import { getPartyDisplayName, isIndividual } from './types';
import type { PartyUnion } from './types';
import './PartyFormPage.css'; // Reusing styles, might need specific selector styles?
import './PartyListPage.css'; // Reusing list styles for table

interface PartySelectorProps {
    selectedPartyId?: string;
    onSelect: (party: PartyUnion) => void;
}

const columnHelper = createColumnHelper<PartyUnion>();

export default function PartySelector({ selectedPartyId, onSelect }: PartySelectorProps) {
    const [searchTerm, setSearchTerm] = useState('');
    const [sorting, setSorting] = useState<SortingState>([]);

    // Fetch parties filtered by search term
    const { data: parties = [], isLoading } = useParties(
        searchTerm ? { givenName: searchTerm, tradingName: searchTerm } : undefined
    );

    const columns = useMemo(
        () => [
            columnHelper.accessor((row) => row['@type'], {
                id: 'type',
                header: 'Type',
                cell: (info) => (
                    <div className={`party-type-badge ${info.getValue().toLowerCase()}`}>
                        {info.getValue() === 'Individual' ? <User size={14} /> : <Building2 size={14} />}
                        {info.getValue()}
                    </div>
                ),
                size: 140,
            }),
            columnHelper.accessor((row) => getPartyDisplayName(row), {
                id: 'name',
                header: 'Name',
                cell: (info) => <span className="party-name">{info.getValue()}</span>,
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
                    selectedPartyId === info.row.original.id && <Check size={18} className="text-success" />
                ),
                size: 50,
            })
        ],
        [selectedPartyId]
    );

    // eslint-disable-next-line react-hooks/incompatible-library
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
        <div className="party-selector-container">
            <div className="search-bar-inline">
                <Search size={18} className="search-icon" />
                <input
                    type="text"
                    placeholder="Search by name..."
                    className="search-input-full"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                />
            </div>

            <div className="party-selector-table-wrapper">
                {isLoading ? (
                    <div className="loading-container">
                        <Loader2 className="spin" size={24} />
                        <p>Loading...</p>
                    </div>
                ) : parties.length === 0 ? (
                    <div className="empty-state">
                        <p>No parties found.</p>
                    </div>
                ) : (
                    <table className="data-table selector-table">
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
                                <tr
                                    key={row.id}
                                    className={`selector-row ${selectedPartyId === row.original.id ? 'selected' : ''}`}
                                    onClick={() => onSelect(row.original)}
                                    role="button"
                                    tabIndex={0}
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter' || e.key === ' ') {
                                            e.preventDefault();
                                            onSelect(row.original);
                                        }
                                    }}
                                >
                                    {row.getVisibleCells().map((cell) => (
                                        <td key={cell.id}>
                                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                        </td>
                                    ))}
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>
        </div>
    );
}
