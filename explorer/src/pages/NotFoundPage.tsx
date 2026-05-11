import { Link } from 'react-router-dom';

export default function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <h1 className="text-4xl font-bold text-white mb-4">404 - Page Not Found</h1>
      <p className="text-gray-400 mb-6">The page you are looking for does not exist.</p>
      <Link to="/" className="text-blue-400 hover:text-blue-300">
        Back to Home
      </Link>
    </div>
  );
}
