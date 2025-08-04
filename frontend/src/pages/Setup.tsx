import {Navigate} from 'react-router';
import {useInitializationStatus} from '../app/InitializationStatusProvider';

export default function Setup() {
    console.log('setup');
    const init = useInitializationStatus();
    
    debugger;
    if (!init.fullStatus) {
        return <div>Loading...</div>;
    }

    if (init.fullStatus?.piholeStatus === "UNINITIALIZED") {
        return <Navigate to='/setup/piholes' replace/>
    }

    return (<Navigate to="/" replace/>);
}