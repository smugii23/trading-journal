import NavBar from "./NavBar";

function Layout({ children }) {
  return (
    <div className="min-h-screen bg-gray-100">
      <NavBar />
      <main className="max-w-4xl mx-auto p-4">{children}</main>
    </div>
  );
}

export default Layout;
