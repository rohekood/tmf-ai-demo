import type { DebugFilterState } from '../types';
import { Search, Filter, Ban } from 'lucide-react';
import './DebugFilters.css';

interface DebugFiltersProps {
    filter: DebugFilterState;
    onChange: (filter: DebugFilterState) => void;
    onClear: () => void;
    totalCount: number;
    filteredCount: number;
}

export function DebugFilters({ filter, onChange, onClear, totalCount, filteredCount }: DebugFiltersProps) {
    const toggleService = (service: string) => {
        const services = filter.services.includes(service)
            ? filter.services.filter((s) => s !== service)
            : [...filter.services, service];
        onChange({ ...filter, services });
    };

    const toggleType = (type: string) => {
        const types = filter.types.includes(type)
            ? filter.types.filter((t) => t !== type)
            : [...filter.types, type];
        onChange({ ...filter, types });
    };

    return (
        <div className="debug-filters card">
            <div className="filter-header">
                <div className="filter-title">
                    <Filter size={16} />
                    <span>Filters</span>
                </div>
                <div className="filter-counts">
                    <span className="count-label">Showing:</span>
                    <span className="count-value">{filteredCount} / {totalCount}</span>
                </div>
            </div>

            <div className="filter-group">
                <div className="search-box">
                    <Search size={16} className="search-icon" />
                    <input
                        type="text"
                        placeholder="Search payload or topic..."
                        value={filter.search}
                        onChange={(e) => onChange({ ...filter, search: e.target.value })}
                        className="search-input"
                    />
                </div>
            </div>

            <div className="filter-group">
                <label>Service:</label>
                <div className="filter-chips">
                    {['party', 'customer', 'bff', 'pocv', 'shopping-cart', 'ordering'].map((service) => (
                        <button
                            key={service}
                            className={`chip ${filter.services.includes(service) ? 'active' : ''}`}
                            onClick={() => toggleService(service)}
                        >
                            {service}
                        </button>
                    ))}
                </div>
            </div>

            <div className="filter-group">
                <label>Type:</label>
                <div className="filter-chips">
                    {['command', 'event', 'query', 'reply'].map((type) => (
                        <button
                            key={type}
                            className={`chip ${filter.types.includes(type) ? 'active' : ''}`}
                            onClick={() => toggleType(type)}
                        >
                            {type}
                        </button>
                    ))}
                </div>
            </div>

            <div className="filter-actions">
                <button className="btn btn-secondary btn-sm btn-full" onClick={onClear}>
                    <Ban size={14} />
                    <span>Clear Messages</span>
                </button>
            </div>
        </div>
    );
}
