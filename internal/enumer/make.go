package enumer

func (g *Generator) makeBitwize() {
	if !g.flags.bitwise {
		return
	}

	g.data.Bitwise = true
}

func (g *Generator) makeJson() {
	if !g.flags.json {
		return
	}
	g.data.Json = true
}

func (g *Generator) makeText() {
	if !g.flags.text {
		return
	}
	g.data.Text = true
}

func (g *Generator) makeSQL() {
	if !g.flags.sql {
		return
	}
	g.data.Sql = true
}
