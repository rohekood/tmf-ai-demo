import './App.css'
import { CustomerList } from './features/customers/CustomerList'

function App() {
  return (
    <>
      <header>
        <h1>TMF Demo Dashboard</h1>
        <p>Managed via Golang BFF & RabbitMQ</p>
      </header>
      <main>
        <section>
          <h2>Customer Management</h2>
          <CustomerList />
        </section>
      </main>
    </>
  )
}

export default App
