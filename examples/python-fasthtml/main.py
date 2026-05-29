from fasthtml.common import *
import asyncio
import sys

if __name__ == "__main__": sys.exit("Run this app with `uvicorn main:app`")

css = Style('''
    body, html { height: 100%; margin: 0; }
    body { display: flex; flex-direction: column; }
    main { flex: 1 0 auto; }
    footer { flex-shrink: 0; padding: 10px; text-align: center; background-color: #333; color: white; }
    footer a { color: #9cf; }
    #grid { display: grid; grid-template-columns: repeat(20, 20px); grid-template-rows: repeat(20, 20px);gap: 1px; }
    .cell { width: 20px; height: 20px; border: 1px solid black; }
    .alive { background-color: green; }
    .dead { background-color: white; }
''')
picolink = Link(
    rel="stylesheet",
    href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css",
)
gridlink = Link(rel="stylesheet", href="https://cdnjs.cloudflare.com/ajax/libs/flexboxgrid/6.3.1/flexboxgrid.min.css", type="text/css")
htmx_ws = Script(src="https://unpkg.com/htmx-ext-ws@2.0.0/ws.js")
app = FastHTML(hdrs=(picolink, gridlink, css, htmx_ws))
rt = app.route

game_state = {'running': False, 'grid': [[0 for _ in range(20)] for _ in range(20)]}
def update_grid(grid: list[list[int]]) -> list[list[int]]:
    new_grid = [[0 for _ in range(20)] for _ in range(20)]
    def count_neighbors(x, y):
        directions = [(-1, -1), (-1, 0), (-1, 1), (0, -1), (0, 1), (1, -1), (1, 0), (1, 1)]
        count = 0
        for dx, dy in directions:
            nx, ny = x + dx, y + dy
            if 0 <= nx < len(grid) and 0 <= ny < len(grid[0]): count += grid[nx][ny]
        return count
    for i in range(len(grid)):
        for j in range(len(grid[0])):
            neighbors = count_neighbors(i, j)
            if grid[i][j] == 1:
                if neighbors < 2 or neighbors > 3: new_grid[i][j] = 0
                else: new_grid[i][j] = 1
            elif neighbors == 3: new_grid[i][j] = 1
    return new_grid

def Grid():
    cells = []
    for y, row in enumerate(game_state['grid']):
        for x, cell in enumerate(row):
            cell_class = 'alive' if cell else 'dead'
            cell = Div(cls=f'cell {cell_class}', hx_put='/update', hx_vals={'x': x, 'y': y}, hx_swap='none', hx_target='#gol', hx_trigger='click')
            cells.append(cell)
    return Div(*cells, id='grid')

def Home():
    gol = Div(Grid(), id='gol', cls='row center-xs')
    run_btn = Button('Run', id='run', cls='col-xs-2', hx_put='/run', hx_target='#gol', hx_swap='none')
    pause_btn = Button('Pause', id='pause', cls='col-xs-2', hx_put='/pause', hx_target='#gol', hx_swap='none')
    reset_btn = Button('Reset', id='reset', cls='col-xs-2', hx_put='/reset', hx_target='#gol', hx_swap='none')
    main = Main(gol, Div(run_btn, pause_btn, reset_btn, cls='row center-xs'), hx_ext="ws", ws_connect="/gol")
    footer = Footer(P('Made by Nathan Cooper. Check out the code', AX('here', href='https://github.com/AnswerDotAI/fasthtml-example/tree/main/game_of_life', target='_blank')))
    return Title('Game of Life'), main, footer

@rt('/')
def get(): return Home()

player_queue = []
async def update_players():
    for i, player in enumerate(player_queue):
        try: await player(Grid())
        except: player_queue.pop(i)
async def on_connect(send): player_queue.append(send)
async def on_disconnect(send): await update_players()

@app.ws('/gol', conn=on_connect, disconn=on_disconnect)
async def ws(msg:str, send): pass

async def background_task():
    while True:
        if game_state['running'] and len(player_queue) > 0:
            game_state['grid'] = update_grid(game_state['grid'])
            await update_players()
        await asyncio.sleep(1.0)

background_task_handle = None

@app.on_event('startup')
async def startup():
    global background_task_handle
    if background_task_handle is None or background_task_handle.done():
        background_task_handle = asyncio.create_task(background_task())

@app.on_event('shutdown')
async def shutdown():
    global background_task_handle
    if background_task_handle is None:
        return
    background_task_handle.cancel()
    try:
        await background_task_handle
    except asyncio.CancelledError:
        pass
    background_task_handle = None

@rt('/update')
async def put(x: int, y: int):
    game_state['grid'][y][x] = 1 if game_state['grid'][y][x] == 0 else 0
    await update_players()

@rt('/run')
async def put():
    game_state['running'] = True
    await update_players()

@rt("/reset")
async def put():
    game_state['grid'] = [[0 for _ in range(20)] for _ in range(20)]
    game_state['running'] = False
    await update_players()

@rt('/pause')
async def put():
    game_state['running'] = False
    await update_players()
