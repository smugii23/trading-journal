import { Link } from 'react-router-dom';

function NavBar() {
    return (
        <nav>
            <ul>
                <li>
                    <Link to="/">Trading Journal</Link>
                </li>
                <li>
                    <Link to="/trades">Trades</Link>
                </li>
                <li>
                    <Link to="/add_trade">Add Trade</Link>
                </li>
                <li>
                    <Link to="/analytics">Analytics</Link>
                </li>
            </ul>
        </nav>
    )
}