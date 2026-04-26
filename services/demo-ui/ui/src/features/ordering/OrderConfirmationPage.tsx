import { useParams, Link } from 'react-router-dom';

export default function OrderConfirmationPage() {
    const { orderId } = useParams<{ orderId: string }>();

    return (
        <div className="p-8 max-w-2xl mx-auto text-center space-y-8 mt-12">
            <div className="inline-flex items-center justify-center w-24 h-24 rounded-full bg-green-100 mb-4">
                <span className="text-4xl text-green-600">✓</span>
            </div>
            <h1 className="text-4xl font-bold text-gray-900">Order Confirmed!</h1>
            <p className="text-xl text-gray-600">Thank you for your purchase.</p>
            <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100 inline-block mt-8">
                <p className="text-gray-500 mb-1">Your Order ID is</p>
                <p className="text-2xl font-mono font-bold text-gray-900">{orderId}</p>
            </div>
            
            <div className="pt-12">
                <Link 
                    to="/"
                    className="text-blue-600 hover:text-blue-800 font-medium"
                >
                    &larr; Return to Dashboard
                </Link>
            </div>
        </div>
    );
}
