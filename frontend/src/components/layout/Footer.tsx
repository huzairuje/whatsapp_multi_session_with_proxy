export default function Footer() {
  return (
    <footer className="bg-white border-t border-gray-200 py-4 px-8">
      <div className="flex items-center justify-between text-sm text-gray-600">
        <div>
          <p>© 2026 WhatsApp Multi-Session Bulk Sender. All rights reserved.</p>
        </div>
        <div className="flex items-center space-x-4">
          <a href="#" className="hover:text-primary transition-colors">
            Documentation
          </a>
          <a href="#" className="hover:text-primary transition-colors">
            Support
          </a>
          <a href="#" className="hover:text-primary transition-colors">
            Privacy Policy
          </a>
        </div>
      </div>
    </footer>
  )
}
