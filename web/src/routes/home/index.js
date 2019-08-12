import { h, Component } from 'preact';
import Card from 'preact-material-components/Card';
import Icon from 'preact-material-components/Icon';
import Fab from 'preact-material-components/Fab';
import LinearProgress from 'preact-material-components/LinearProgress';
import 'preact-material-components/Card/style.css';
import 'preact-material-components/Icon/style.css';
import 'preact-material-components/Fab/style.css';
import 'preact-material-components/LinearProgress/style.css';
import style from './style';
import { getDays, setDays, getToken, setToken, refreshWithToken } from '../../utils/DedoFEUPService'
import { route } from 'preact-router';

export default class Home extends Component {
	state = {
		loading: false
	}

	refresh = () => {
		this.setState({ loading: true })
		refreshWithToken(getToken()).then((data) => {
			this.setState({ loading: false })
			setDays(data.Days)
		}).catch((error) => {
			this.setState({ loading: false })
			if (error.response && error.response.data) {
				if(error.response.data.Error === "not logged in") {
					setToken(null)
					route("/login/", true)
					return
				}
				else if(error.response.data.Error === "invalid token") {
					setToken(null)
					route("/login/", true)
					return
				}
			}
			alert('unexpected error')
		});
	}

	render(props, state) {
		if (state.loading) {
			return (
				<div class={`${style.home} page`}>
				<h4>Refreshing...</h4>
					<LinearProgress indeterminate />
				</div>
			)
		} else {
			return (<div class={`${style.home} page`}>
				<Fab ripple class={style.fab} onClick={this.refresh}><Fab.Icon>refresh</Fab.Icon></Fab>
				<h1></h1>
				{ getDays().map(item => 
					<div>
						<Card>
							<div class={style.cardHeader}>
								<h2 class=" mdc-typography--title">{item.Date}</h2>
								<div class=" mdc-typography--caption">
									<div>
									<Icon>alarm</Icon> {item.MorningIn} / {item.MorningOut}
									</div>
									<div>
									<Icon>wb_sunny</Icon> {item.AfternoonIn} / {item.AfternoonOut}
									</div>
								</div>
							</div>
							</Card>
					</div>
				)}
			</div>
		)
		}
	}
}
