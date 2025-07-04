package foo20

import (
	"std"

	"gno.land/p/demo/avl"
	p "gno.land/p/demo/avl/pager"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/demo/grc20reg"
)

var (
	instances avl.Tree // symbol -> *instance
	pager     = p.NewPager(rotree.Wrap(&instances, nil), 20, false)
)

type instance struct {
	token  *grc20.Token
	ledger *grc20.PrivateLedger
	admin  *ownable.Ownable
	faucet int64 // per-request amount. disabled if 0.
}

func New(cur realm, name, symbol string, decimals int, initialMint, faucet int64) {
	caller := std.PreviousRealm().Address()
	NewWithAdmin(cur, name, symbol, decimals, initialMint, faucet, caller)
}

func NewWithAdmin(cur realm, name, symbol string, decimals int, initialMint, faucet int64, admin std.Address) {
	exists := instances.Has(symbol)
	if exists {
		panic("token already exists")
	}

	token, ledger := grc20.NewToken(name, symbol, decimals)
	if initialMint > 0 {
		ledger.Mint(admin, initialMint)
	}

	inst := instance{
		token:  token,
		ledger: ledger,
		admin:  ownable.NewWithAddress(admin),
		faucet: faucet,
	}
	instances.Set(symbol, &inst)

	grc20reg.Register(cross, token, symbol)
}

func Bank(symbol string) *grc20.Token {
	inst := mustGetInstance(symbol)
	return inst.token
}

func TotalSupply(symbol string) int64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().TotalSupply()
}

func HasAddr(symbol string, owner std.Address) bool {
	inst := mustGetInstance(symbol)
	return inst.token.HasAddr(owner)
}

func BalanceOf(symbol string, owner std.Address) int64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().BalanceOf(owner)
}

func Allowance(symbol string, owner, spender std.Address) int64 {
	inst := mustGetInstance(symbol)
	return inst.token.ReadonlyTeller().Allowance(owner, spender)
}

func Transfer(cur realm, symbol string, to std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.Transfer(to, amount))
}

func Approve(cur realm, symbol string, spender std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.Approve(spender, amount))
}

func TransferFrom(cur realm, symbol string, from, to std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	caller := std.PreviousRealm().Address()
	teller := inst.ledger.ImpersonateTeller(caller)
	checkErr(teller.TransferFrom(from, to, amount))
}

// faucet.
func Faucet(cur realm, symbol string) {
	inst := mustGetInstance(symbol)
	if inst.faucet == 0 {
		panic("faucet disabled for this token")
	}
	// FIXME: add limits?
	// FIXME: add payment in gnot?
	caller := std.PreviousRealm().Address()
	checkErr(inst.ledger.Mint(caller, inst.faucet))
}

func Mint(cur realm, symbol string, to std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertOwnedByPrevious()
	checkErr(inst.ledger.Mint(to, amount))
}

func Burn(cur realm, symbol string, from std.Address, amount int64) {
	inst := mustGetInstance(symbol)
	inst.admin.AssertOwnedByPrevious()
	checkErr(inst.ledger.Burn(from, amount))
}

// instance admin functionality
func DropInstanceOwnership(cur realm, symbol string) {
	inst := mustGetInstance(symbol)
	checkErr(inst.admin.DropOwnershipByCurrent())
}

func TransferInstanceOwnership(cur realm, symbol string, newOwner std.Address) {
	inst := mustGetInstance(symbol)
	checkErr(inst.admin.TransferOwnership(newOwner))
}

func ListTokens(pageNumber, pageSize int) []*grc20.Token {
	page := pager.GetPageWithSize(pageNumber, pageSize)

	tokens := make([]*grc20.Token, len(page.Items))
	for i := range page.Items {
		tokens[i] = page.Items[i].Value.(*instance).token
	}

	return tokens
}

func Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", renderHome)
	router.HandleFunc("{symbol}", renderToken)
	router.HandleFunc("{symbol}/balance/{address}", renderBalance)
	return router.Render(path)
}

func renderHome(res *mux.ResponseWriter, req *mux.Request) {
	out := md.H1(ufmt.Sprintf("GRC20 Tokens (%d)", instances.Size()))

	// Get the current page of tokens based on the request path.
	page := pager.MustGetPageByPath(req.RawPath)

	// Render the list of tokens.
	for _, item := range page.Items {
		token := item.Value.(*instance).token
		out += md.BulletItem(
			md.Link(
				ufmt.Sprintf("%s ($%s)", token.GetName(), token.GetSymbol()),
				ufmt.Sprintf("/r/demo/grc20factory:%s", token.GetSymbol()),
			),
		)
	}
	out += "\n"

	// Add the page picker.
	out += md.Paragraph(page.Picker(req.Path))

	res.Write(out)
}

func renderToken(res *mux.ResponseWriter, req *mux.Request) {
	// Get the token symbol from the request.
	symbol := req.GetVar("symbol")
	inst := mustGetInstance(symbol)

	// Render the token details.
	out := inst.token.RenderHome()
	out += md.BulletItem(
		ufmt.Sprintf("%s: %s", md.Bold("Admin"), inst.admin.Owner()),
	)

	res.Write(out)
}

func renderBalance(res *mux.ResponseWriter, req *mux.Request) {
	var (
		symbol = req.GetVar("symbol")
		addr   = req.GetVar("address")
	)

	// Get the balance of the specified address for the token.
	inst := mustGetInstance(symbol)
	balance := inst.token.CallerTeller().BalanceOf(std.Address(addr))

	// Render the balance information.
	out := md.Paragraph(
		ufmt.Sprintf("%s balance: %d", md.Bold(addr), balance),
	)

	res.Write(out)
}

func mustGetInstance(symbol string) *instance {
	t, exists := instances.Get(symbol)
	if !exists {
		panic("token instance does not exist")
	}
	return t.(*instance)
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
